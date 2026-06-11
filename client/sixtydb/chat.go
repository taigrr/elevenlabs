package sixtydb

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultLLMModel is 60db's small fast model — override per call.
const DefaultLLMModel = "60db-tiny"

// Message is the OpenAI-compatible chat message shape that 60db's
// /v1/chat/completions endpoint accepts unmodified.
type Message struct {
	Role    string `json:"role"`    // "system" | "user" | "assistant"
	Content string `json:"content"`
}

type chatRequest struct {
	Model              string                 `json:"model"`
	Messages           []Message              `json:"messages"`
	Stream             bool                   `json:"stream"`
	TopK               int                    `json:"top_k,omitempty"`
	Temperature        float64                `json:"temperature,omitempty"`
	MaxTokens          int                    `json:"max_tokens,omitempty"`
	ChatTemplateKwargs map[string]interface{} `json:"chat_template_kwargs,omitempty"`
}

type chatChoice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
	Delta   Message `json:"delta"` // only populated in stream mode
}

type chatResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
}

// ChatCompletionOptions mirrors the small subset of OpenAI parameters
// 60db documents. All fields are optional; zero values fall through to
// 60db defaults via omitempty.
type ChatCompletionOptions struct {
	TopK        int
	Temperature float64
	MaxTokens   int
}

// ChatCompletion posts to POST /v1/chat/completions and returns the
// assistant's reply content.
//
// Reference: https://docs.60db.ai/api-reference/llm/chat-completion
func (c Client) ChatCompletion(ctx context.Context, model string, messages []Message, opts ChatCompletionOptions) (string, error) {
	if model == "" {
		model = DefaultLLMModel
	}
	body, _ := json.Marshal(chatRequest{
		Model:              model,
		Messages:           messages,
		Stream:             false,
		TopK:               opts.TopK,
		Temperature:        opts.Temperature,
		MaxTokens:          opts.MaxTokens,
		ChatTemplateKwargs: map[string]interface{}{"enable_thinking": false},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs/client/sixtydb")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusUnauthorized:
		return "", ErrUnauthorized
	case http.StatusOK:
		// fall through
	default:
		preview, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return "", fmt.Errorf("60db /v1/chat/completions %d: %s", res.StatusCode, string(preview))
	}

	var data chatResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data.Choices) == 0 {
		return "", fmt.Errorf("60db /v1/chat/completions returned no choices")
	}
	return data.Choices[0].Message.Content, nil
}

// ChatCompletionStream posts with stream:true and writes the assistant's
// delta tokens to w as they arrive. Matches the SSE-pass-through pattern
// the parent ElevenLabs package uses for its streaming TTS surface.
//
// The upstream emits OpenAI-style SSE: lines prefixed `data: ` followed by
// a JSON payload, terminated by `data: [DONE]`.
func (c Client) ChatCompletionStream(ctx context.Context, w io.Writer, model string, messages []Message, opts ChatCompletionOptions) error {
	if model == "" {
		model = DefaultLLMModel
	}
	body, _ := json.Marshal(chatRequest{
		Model:              model,
		Messages:           messages,
		Stream:             true,
		TopK:               opts.TopK,
		Temperature:        opts.Temperature,
		MaxTokens:          opts.MaxTokens,
		ChatTemplateKwargs: map[string]interface{}{"enable_thinking": false},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs/client/sixtydb")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusOK:
		// fall through
	default:
		preview, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("60db /v1/chat/completions %d: %s", res.StatusCode, string(preview))
	}

	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			return nil
		}
		var data chatResponse
		if err := json.Unmarshal([]byte(payload), &data); err != nil {
			continue
		}
		if len(data.Choices) == 0 {
			continue
		}
		delta := data.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		if _, err := w.Write([]byte(delta)); err != nil {
			return err
		}
	}
	return scanner.Err()
}
