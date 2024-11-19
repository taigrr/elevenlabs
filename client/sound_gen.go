package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/supagroova/elevenlabs/client/types"
)

// SoundGenerationWriter generates a sound effect from text and writes it to the provided writer.
// If durationSeconds is 0, it will be omitted from the request and the API will determine the optimal duration.
// If promptInfluence is 0, it will default to 0.3.
func (c Client) SoundGenerationWriter(ctx context.Context, w io.Writer, text string, durationSeconds, promptInfluence float64) error {
	params := types.SoundGeneration{
		Text:            text,
		PromptInfluence: 0.3, // default value
	}
	
	if promptInfluence != 0 {
		params.PromptInfluence = promptInfluence
	}
	if durationSeconds != 0 {
		params.DurationSeconds = durationSeconds
	}

	body, err := c.requestSoundGeneration(ctx, params)
	if err != nil {
		return err
	}
	defer body.Close()
	_, err = io.Copy(w, body)
	return err
}

// SoundGeneration generates a sound effect from text and returns the audio as bytes.
// If durationSeconds is 0, it will be omitted from the request and the API will determine the optimal duration.
// If promptInfluence is 0, it will default to 0.3.
func (c Client) SoundGeneration(ctx context.Context, text string, durationSeconds, promptInfluence float64) ([]byte, error) {
	params := types.SoundGeneration{
		Text:            text,
		PromptInfluence: 0.3, // default value
	}
	
	if promptInfluence != 0 {
		params.PromptInfluence = promptInfluence
	}
	if durationSeconds != 0 {
		params.DurationSeconds = durationSeconds
	}

	body, err := c.requestSoundGeneration(ctx, params)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	
	var b bytes.Buffer
	_, err = io.Copy(&b, body)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (c Client) requestSoundGeneration(ctx context.Context, params types.SoundGeneration) (io.ReadCloser, error) {
	url := c.endpoint + "/v1/sound-generation"
	client := &http.Client{}
	
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/supagroova/elevenlabs")
	req.Header.Set("accept", "audio/mpeg")
	
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	
	return res.Body, nil
}
