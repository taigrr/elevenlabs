package sixtydb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// SpeechToTextRequest mirrors the parent package's surface so caller
// code that wraps both providers can share one config struct shape.
type SpeechToTextRequest struct {
	Language string // ISO-639-1, "" = auto-detect
	Diarize  bool
	Keywords string // CSV, optional vocabulary boost
}

// SpeechToTextResponse is a flat view of 60db's /stt response. The full
// segments/words tree is left as raw JSON so callers can decode it on
// demand without bloating this struct.
type SpeechToTextResponse struct {
	RequestID    string          `json:"request_id"`
	Text         string          `json:"text"`
	Language     string          `json:"language"`
	LanguageName string          `json:"language_name"`
	DurationSec  float64         `json:"duration_sec"`
	Segments     json.RawMessage `json:"segments,omitempty"`
	Words        json.RawMessage `json:"words,omitempty"`
}

// ConvertSpeechToText reads `audioFilePath` and uploads it to /stt.
// Mirrors the parent package's signature one-for-one.
//
// Reference: https://docs.60db.ai/api-reference/stt/speech-to-text
func (c Client) ConvertSpeechToText(ctx context.Context, audioFilePath string, request SpeechToTextRequest) (*SpeechToTextResponse, error) {
	file, err := os.Open(audioFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()
	return c.ConvertSpeechToTextFromReader(ctx, file, filepath.Base(audioFilePath), request)
}

// ConvertSpeechToTextFromReader streams audio bytes into a multipart
// body and posts to /stt. Same shape as the elevenlabs client's reader
// variant so swapping providers requires only an import change.
func (c Client) ConvertSpeechToTextFromReader(ctx context.Context, reader io.Reader, filename string, request SpeechToTextRequest) (*SpeechToTextResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = io.Copy(part, reader); err != nil {
		return nil, fmt.Errorf("failed to copy audio data: %w", err)
	}
	if request.Language != "" {
		_ = writer.WriteField("language", request.Language)
	}
	if request.Diarize {
		_ = writer.WriteField("diarize", "true")
	}
	if request.Keywords != "" {
		_ = writer.WriteField("keywords", request.Keywords)
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/stt", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs/client/sixtydb")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case http.StatusOK:
		var sttResp SpeechToTextResponse
		if err := json.NewDecoder(res.Body).Decode(&sttResp); err != nil {
			return nil, fmt.Errorf("failed to parse API response: %w", err)
		}
		return &sttResp, nil
	default:
		preview, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return nil, fmt.Errorf("60db /stt %d: %s", res.StatusCode, string(preview))
	}
}
