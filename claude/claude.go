package claude

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

const (
	ApiEndpoint = "https://api.anthropic.com/v1/messages"
	model       = "claude-3-5-sonnet-20240620"
	transcript  = "claude.json"
)

var apiKey = os.Getenv("ANTHROPIC_API_KEY")

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ApiRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type ApiResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason   string  `json:"stop_reason"`
	StopSequence *string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func Completions(ctx context.Context, role string, fake bool, t *template.Template, context any) (*ApiResponse, error) {
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

	messages := []Message{
		{Role: "user", Content: sb.String()},
	}

	apiRequest := ApiRequest{
		Model:     model,
		MaxTokens: 1024,
		Messages:  messages,
	}

	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ApiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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

	var apiResponse ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := appendTranscript(RequestResponse{
		Role:     role,
		Request:  apiRequest,
		Response: apiResponse,
	}); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func parseJsonResponse1[T any](choice string, result T) error {
	if err := json.Unmarshal([]byte(choice), result); err != nil {
		return err
	}
	return nil
}
func parseJsonResponse2[T any](choice string, result T) error {
	text := choice
	i := strings.Index(choice, "[")
	if i == -1 {
		return errors.Errorf("could not find boundary")
	}
	text = choice[i+1:]
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

func parseJsonResponse3[T any](choice string, result T) error {
	// XXX: use re.
	text := choice
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
func parseJsonResponse4[T any](choice string, result T) error {
	var response struct {
		Content T `json:"search_text"`
	}
	if err := json.Unmarshal([]byte(choice), &response); err != nil {
		return err
	}
	result = response.Content
	return nil
}
func ParseJsonResponse[T any](choice string, result T) error {
	if err := parseJsonResponse1(choice, result); err == nil {
		return nil
	}
	if err := parseJsonResponse2(choice, result); err == nil {
		return nil
	}
	if err := parseJsonResponse3(choice, result); err == nil {
		return nil
	}
	if err := parseJsonResponse4(choice, result); err == nil {
		return nil
	}
	return errors.New("could not parse json response")
}

type RequestResponse struct {
	Role     string      `json:"role"`
	Request  ApiRequest  `json:"request"`
	Response ApiResponse `json:"response"`
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
