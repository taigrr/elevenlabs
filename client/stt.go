package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/taigrr/elevenlabs/client/types"
)

// ConvertSpeechToText converts audio to text using the specified file path
func (c *Client) ConvertSpeechToText(ctx context.Context, audioFilePath string, request types.SpeechToTextRequest) (*types.SpeechToTextResponse, error) {
	file, err := os.Open(audioFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	return c.ConvertSpeechToTextFromReader(ctx, file, filepath.Base(audioFilePath), request)
}

// ConvertSpeechToTextFromReader converts audio to text using the provided reader
func (c *Client) ConvertSpeechToTextFromReader(ctx context.Context, reader io.Reader, filename string, request types.SpeechToTextRequest) (*types.SpeechToTextResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("model_id", string(request.ModelID)); err != nil {
		return nil, fmt.Errorf("failed to write model_id field: %w", err)
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err = io.Copy(part, reader); err != nil {
		return nil, fmt.Errorf("failed to copy audio data: %w", err)
	}

	if request.LanguageCode != "" {
		if err := writer.WriteField("language_code", request.LanguageCode); err != nil {
			return nil, fmt.Errorf("failed to write language_code field: %w", err)
		}
	}

	if request.NumSpeakers != 0 {
		if err := writer.WriteField("num_speakers", fmt.Sprintf("%d", request.NumSpeakers)); err != nil {
			return nil, fmt.Errorf("failed to write num_speakers field: %w", err)
		}
	}
	if request.TagAudioEvents {
		if err := writer.WriteField("tag_audio_events", "true"); err != nil {
			return nil, fmt.Errorf("failed to write tag_audio_events field: %w", err)
		}
	}
	if request.TimestampsGranularity != "" {
		if err := writer.WriteField("timestamps_granularity", string(request.TimestampsGranularity)); err != nil {
			return nil, fmt.Errorf("failed to write timestamps_granularity field: %w", err)
		}
	}
	if request.Diarize {
		if err := writer.WriteField("diarize", "true"); err != nil {
			return nil, fmt.Errorf("failed to write diarize field: %w", err)
		}
	}

	if err = writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	client := &http.Client{}
	url := fmt.Sprintf(c.endpoint + "/v1/speech-to-text")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("xi-api-key", c.apiKey)

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	switch res.StatusCode {
	case 401:
		return nil, ErrUnauthorized
	case 200:
		var sttResponse types.SpeechToTextResponse
		if err := json.NewDecoder(res.Body).Decode(&sttResponse); err != nil {
			return nil, fmt.Errorf("failed to parse API response: %w", err)
		}

		return &sttResponse, nil
	case 422:
		fallthrough
	default:
		ve := types.ValidationError{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&ve)
		if jerr != nil {
			err = errors.Join(err, jerr)
		} else {
			err = errors.Join(err, ve)
		}
		return nil, err
	}
}
