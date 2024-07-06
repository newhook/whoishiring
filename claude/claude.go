package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
)

const (
	ApiEndpoint = "https://api.anthropic.com/v1/messages"
)

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

const model = "claude-3-5-sonnet-20240620"

func Completions(ctx context.Context, file string, call bool, t *template.Template, context any) (*ApiResponse, error) {
	sb := &strings.Builder{}
	err := t.Execute(sb, context)
	if err != nil {
		return nil, err
	}
	fmt.Println(sb.String())
	var body []byte
	if call {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		client := &http.Client{}

		messages := []Message{
			{Role: "user", Content: sb.String()},
		}

		requestBody, err := json.Marshal(ApiRequest{
			Model:     model,
			MaxTokens: 1024,
			Messages:  messages,
		})
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, "POST", ApiEndpoint, bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(file, body, 0644)
		if err != nil {
			return nil, err
		}

	} else {
		var err error
		body, err = os.ReadFile(file)
		if err != nil {
			return nil, err
		}
	}
	var apiResponse ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
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
		return fmt.Errorf("could not find boundary")
	}
	text = choice[i+1:]
	i = strings.LastIndex(text, "]")
	if i == -1 {
		return fmt.Errorf("could not find boundary")
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
		return fmt.Errorf("could not find boundary")
	}
	text = text[i+3:]
	i = strings.Index(text, boundary)
	if i == -1 {
		return fmt.Errorf("could not find boundary")
	}
	text = text[:i]
	fmt.Println(text)
	if !strings.HasPrefix(text, "json") {
		return fmt.Errorf("expected json")
	}
	text = text[4:]
	if err := json.Unmarshal([]byte(text), result); err != nil {
		return err
	}
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
	return errors.New("could not parse json response")
}
