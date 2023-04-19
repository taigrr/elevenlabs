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
	"strings"

	"github.com/taigrr/elevenlabs/client/types"
)

func (c Client) CreateVoice(ctx context.Context, name, description string, labels []string, files []*os.File) error {
	url := c.endpoint + "/v1/voices/add"
	client := &http.Client{}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for _, r := range files {
		var fw io.Writer
		var err error
		if fw, err = w.CreateFormFile("files[]", r.Name()); err != nil {
			return err
		}
		if _, err := io.Copy(fw, r); err != nil {
			return err
		}
	}
	w.WriteField("name", name)
	w.WriteField("description", description)
	w.WriteField("name", strings.Join(labels, ", "))
	w.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "application/json")
	res, err := client.Do(req)
	switch res.StatusCode {
	case 422:
		ve := types.ValidationError{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&ve)
		if jerr != nil {
			err = errors.Join(err, jerr)
		} else {
			err = errors.Join(err, ve)
		}
		return err
	case 401:
		return ErrUnauthorized
	case 200:
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.Join(err, ErrUnspecified)
	}
}

func (c Client) DeleteVoice(ctx context.Context, voiceID string) error {
	url := fmt.Sprintf(c.endpoint+"/v1/voices/%s", voiceID)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "application/json")
	res, err := client.Do(req)
	switch res.StatusCode {
	case 422:
		ve := types.ValidationError{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&ve)
		if jerr != nil {
			err = errors.Join(err, jerr)
		} else {
			err = errors.Join(err, ve)
		}
		return err
	case 401:
		return ErrUnauthorized
	case 200:
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.Join(err, ErrUnspecified)
	}
}

func (c Client) EditVoiceSettings(ctx context.Context, voiceID string, settings types.SynthesisOptions) error {
	url := fmt.Sprintf(c.endpoint+"/v1/voices/%s/settings/edit", voiceID)
	client := &http.Client{}
	b, _ := json.Marshal(settings)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "application/json")
	res, err := client.Do(req)
	switch res.StatusCode {
	case 422:
		ve := types.ValidationError{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&ve)
		if jerr != nil {
			err = errors.Join(err, jerr)
		} else {
			err = errors.Join(err, ve)
		}
		return err
	case 401:
		return ErrUnauthorized
	case 200:
		if err != nil {
			return err
		}
		so := types.SynthesisOptions{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&so)
		if jerr != nil {
			return jerr
		}
		return nil
	default:
		return errors.Join(err, ErrUnspecified)
	}
}

func (c Client) EditVoice(ctx context.Context, voiceID, name, description string, labels []string, files []*os.File) error {
	url := fmt.Sprintf(c.endpoint+"/v1/voices/%s/edit", voiceID)
	client := &http.Client{}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for _, r := range files {
		var fw io.Writer
		var err error
		if fw, err = w.CreateFormFile("files[]", r.Name()); err != nil {
			return err
		}
		if _, err := io.Copy(fw, r); err != nil {
			return err
		}
	}
	w.WriteField("name", name)
	w.WriteField("description", description)
	w.WriteField("name", strings.Join(labels, ", "))
	w.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "application/json")
	res, err := client.Do(req)
	switch res.StatusCode {
	case 422:
		ve := types.ValidationError{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&ve)
		if jerr != nil {
			err = errors.Join(err, jerr)
		} else {
			err = errors.Join(err, ve)
		}
		return err
	case 401:
		return ErrUnauthorized
	case 200:
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.Join(err, ErrUnspecified)
	}
}

func (c Client) defaultVoiceSettings(ctx context.Context) (types.SynthesisOptions, error) {
	url := c.endpoint + "/v1/voices/settings/default"
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return types.SynthesisOptions{}, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "application/json")
	res, err := client.Do(req)
	switch res.StatusCode {
	case 422:
		ve := types.ValidationError{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&ve)
		if jerr != nil {
			err = errors.Join(err, jerr)
		} else {
			err = errors.Join(err, ve)
		}
		return types.SynthesisOptions{}, err
	case 401:
		return types.SynthesisOptions{}, ErrUnauthorized
	case 200:
		if err != nil {
			return types.SynthesisOptions{}, err
		}
		so := types.SynthesisOptions{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&so)
		if jerr != nil {
			return types.SynthesisOptions{}, jerr
		}
		return so, nil
	default:
		return types.SynthesisOptions{}, errors.Join(err, ErrUnspecified)
	}
}

func (c Client) GetVoiceSettings(ctx context.Context, voiceID string) (types.SynthesisOptions, error) {
	url := fmt.Sprintf(c.endpoint+"/v1/voices/%s/settings", voiceID)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return types.SynthesisOptions{}, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "application/json")
	res, err := client.Do(req)
	switch res.StatusCode {
	case 422:
		ve := types.ValidationError{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&ve)
		if jerr != nil {
			err = errors.Join(err, jerr)
		} else {
			err = errors.Join(err, ve)
		}
		return types.SynthesisOptions{}, err
	case 401:
		return types.SynthesisOptions{}, ErrUnauthorized
	case 200:
		if err != nil {
			return types.SynthesisOptions{}, err
		}
		so := types.SynthesisOptions{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&so)
		if jerr != nil {
			return types.SynthesisOptions{}, jerr
		}
		return so, nil
	default:
		return types.SynthesisOptions{}, errors.Join(err, ErrUnspecified)
	}
}

func (c Client) GetVoice(ctx context.Context, voiceID string) (types.VoiceResponseModel, error) {
	url := fmt.Sprintf(c.endpoint+"/v1/voices/%s", voiceID)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return types.VoiceResponseModel{}, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "application/json")
	res, err := client.Do(req)
	switch res.StatusCode {
	case 422:
		ve := types.ValidationError{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&ve)
		if jerr != nil {
			err = errors.Join(err, jerr)
		} else {
			err = errors.Join(err, ve)
		}
		return types.VoiceResponseModel{}, err
	case 401:
		return types.VoiceResponseModel{}, ErrUnauthorized
	case 200:
		if err != nil {
			return types.VoiceResponseModel{}, err
		}

		vrm := types.VoiceResponseModel{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&vrm)
		if jerr != nil {
			return types.VoiceResponseModel{}, jerr
		}
		return vrm, nil
	default:
		return types.VoiceResponseModel{}, errors.Join(err, ErrUnspecified)
	}
}

func (c Client) GetVoices(ctx context.Context) ([]types.VoiceResponseModel, error) {
	url := c.endpoint + "/v1/voices"
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []types.VoiceResponseModel{}, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "application/json")
	res, err := client.Do(req)
	switch res.StatusCode {
	case 422:
		ve := types.ValidationError{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&ve)
		if jerr != nil {
			err = errors.Join(err, jerr)
		} else {
			err = errors.Join(err, ve)
		}
		return []types.VoiceResponseModel{}, err
	case 401:
		return []types.VoiceResponseModel{}, ErrUnauthorized
	case 200:
		if err != nil {
			return []types.VoiceResponseModel{}, err
		}
		vr := types.GetVoicesResponseModel{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&vr)
		if jerr != nil {
			return []types.VoiceResponseModel{}, jerr
		}
		return vr.Voices, nil

	default:
		return []types.VoiceResponseModel{}, errors.Join(err, ErrUnspecified)
	}
}

func (c Client) GetVoiceIDs(ctx context.Context) ([]string, error) {
	rms, err := c.GetVoices(ctx)
	if err != nil {
		return []string{}, err
	}
	list := []string{}

	for _, v := range rms {
		list = append(list, v.VoiceID)
	}

	return list, nil
}
