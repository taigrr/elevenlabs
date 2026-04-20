package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return res.Body, nil
}
