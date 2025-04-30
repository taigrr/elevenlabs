package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/taigrr/elevenlabs/client/types"
)

func WithPreviousText(previousText string) types.TTSParam {
	return func(tts *types.TTS) {
		tts.PreviousText = previousText
	}
}

func WithNextText(nextText string) types.TTSParam {
	return func(tts *types.TTS) {
		tts.NextText = nextText
	}
}

func (c Client) TTSWriter(ctx context.Context, w io.Writer, text, modelID, voiceID string, options types.SynthesisOptions, optionalParams ...types.TTSParam) error {
	params := types.TTS{
		Text:    text,
		VoiceID: voiceID,
		ModelID: modelID,
	}
	for _, p := range optionalParams {
		p(&params)
	}

	body, err := c.requestTTS(ctx, params, options)
	if err != nil {
		return err
	}
	defer body.Close()
	io.Copy(w, body)
	return nil
}

func (c Client) TTS(ctx context.Context, text, voiceID, modelID string, options types.SynthesisOptions, optionalParams ...types.TTSParam) ([]byte, error) {
	params := types.TTS{
		Text:    text,
		VoiceID: voiceID,
		ModelID: modelID,
	}
	for _, p := range optionalParams {
		p(&params)
	}

	body, err := c.requestTTS(ctx, params, options)
	if err != nil {
		return []byte{}, err
	}
	defer body.Close()
	b := bytes.Buffer{}
	io.Copy(&b, body)
	return b.Bytes(), nil
}

func (c Client) TTSStream(ctx context.Context, w io.Writer, text, voiceID string, options types.SynthesisOptions, optionalParams ...types.TTSParam) error {
	params := types.TTS{
		Text:    text,
		VoiceID: voiceID,
		Stream:  true,
	}
	for _, p := range optionalParams {
		p(&params)
	}

	body, err := c.requestTTS(ctx, params, options)
	if err != nil {
		return err
	}
	defer body.Close()
	io.Copy(w, body)
	return nil
}

func (c Client) requestTTS(ctx context.Context, params types.TTS, options types.SynthesisOptions) (io.ReadCloser, error) {
	options.Clamp()
	url := fmt.Sprintf(c.endpoint+"/v1/text-to-speech/%s", params.VoiceID)
	if params.Stream {
		url += "/stream"
	}
	b, _ := json.Marshal(params)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "audio/mpeg")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch res.StatusCode {
	case 401:
		return nil, ErrUnauthorized
	case 200:
		return res.Body, nil
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
