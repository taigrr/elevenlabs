package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/taigrr/elevenlabs/client/types"
)

func (c Client) SoundGeneration(ctx context.Context, text string, durationSeconds, promptInfluence float64) (io.ReadCloser, error) {
	params := types.SoundGeneration{
		Text:            text,
		PromptInfluence: 0.3,
	}

	if promptInfluence != 0 {
		params.PromptInfluence = promptInfluence
	}
	if durationSeconds != 0 {
		params.DurationSeconds = durationSeconds
	}

	return c.requestSoundGeneration(ctx, params)
}

func (c Client) requestSoundGeneration(ctx context.Context, params types.SoundGeneration) (io.ReadCloser, error) {
	url := c.endpoint + "/v1/sound-generation"
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "audio/mpeg")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		return res.Body, nil
	case http.StatusUnauthorized:
		res.Body.Close()
		return nil, ErrUnauthorized
	case http.StatusUnprocessableEntity:
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
