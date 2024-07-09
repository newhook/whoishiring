package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
)

const transcript = "openai.json"
const url = "https://api.openai.com/v1/chat/completions"

var (
	apiKey = os.Getenv("OPENAI_API_KEY")
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	Logprobs     *string `json:"logprobs"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int      `json:"created"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}

func Completions(ctx context.Context, role string, fake bool, t *template.Template, context any) (*ChatResponse, error) {
	if fake {
		last, err := readLast(role)
		if err != nil {
			return nil, err
		}
		return &last.Response, nil
	}

	sb := &strings.Builder{}
	err := t.Execute(sb, context)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	chatRequest := ChatRequest{
		Model: "gpt-4o",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are job search assistant. Don't explain anything. Provide all results in json",
			},
			{
				Role:    "user",
				Content: sb.String(),
			},
		},
	}

	jsonData, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response from the embedding API: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var cr ChatResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, err
	}

	if err := appendTranscript(RequestResponse{
		Role:     role,
		Request:  chatRequest,
		Response: cr,
	}); err != nil {
		return nil, err
	}

	return &cr, nil
}
func parseJsonResponse1[T any](choice Choice, result T) error {
	if err := json.Unmarshal([]byte(choice.Message.Content), result); err != nil {
		return err
	}
	return nil
}
func parseJsonResponse2[T any](choice Choice, result T) error {
	text := choice.Message.Content
	i := strings.Index(choice.Message.Content, "[")
	if i == -1 {
		return errors.Errorf("could not find boundary")
	}
	text = choice.Message.Content[i+1:]
	i = strings.LastIndex(text, "]")
	if i == -1 {
		return errors.Errorf("could not find boundary")
	}
	text = text[:i]
	if err := json.Unmarshal([]byte("["+text+"]"), result); err == nil {
		return err
	}
	return nil
}

func parseJsonResponse3[T any](choice Choice, result T) error {
	// XXX: use re.
	text := choice.Message.Content
	boundary := "```"
	i := strings.Index(text, boundary)
	if i == -1 {
		return errors.Errorf("could not find boundary")
	}
	text = text[i+3:]
	i = strings.Index(text, boundary)
	if i == -1 {
		return errors.Errorf("could not find boundary")
	}
	text = text[:i]
	fmt.Println(text)
	if !strings.HasPrefix(text, "json") {
		return errors.Errorf("expected json")
	}
	text = text[4:]
	if err := json.Unmarshal([]byte(text), result); err != nil {
		return err
	}
	return nil
}

func ParseJsonResponse[T any](choice Choice, result T) error {
	if err := parseJsonResponse1(choice, result); err == nil {
		return nil
	}
	if err := parseJsonResponse2(choice, result); err == nil {
		return nil
	}
	if err := parseJsonResponse3(choice, result); err == nil {
		return nil
	}
	return errors.New("could not parse json response")
}

type RequestResponse struct {
	Role     string       `json:"role"`
	Request  ChatRequest  `json:"request"`
	Response ChatResponse `json:"response"`
}

func readLast(role string) (RequestResponse, error) {
	file, err := os.Open(transcript)
	if err != nil {
		return RequestResponse{}, errors.WithStack(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var last RequestResponse
	for decoder.More() {
		var obj RequestResponse
		err := decoder.Decode(&obj)
		if err != nil {
			return RequestResponse{}, errors.WithStack(err)
		}
		if obj.Role == role {
			last = obj
		}
	}
	return last, nil
}

func appendTranscript(r RequestResponse) error {
	b, err := json.Marshal(r)
	if err != nil {
		return errors.WithStack(err)
	}
	file, err := os.OpenFile(transcript, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(b); err != nil {
		return errors.WithStack(err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
