package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/taigrr/elevenlabs/client/types"
)

func (c Client) CreateVoice(ctx context.Context, name string, files []*os.File, description string, labels string) error {
	localVarHttpMethod = strings.ToUpper("Post")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/add"
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

func (c Client) EditVoiceSettings(ctx context.Context, voiceID string, types.SynthesisOptions) error {
	url := fmt.Sprintf(c.endpoint+"/v1/voices/%s/settings/edit", voiceID)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
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

func (c Client) EditVoice(ctx context.Context, voiceID, name, description string,  files []*os.File, labels []string ) error {
	localVarHttpMethod = strings.ToUpper("Post")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/edit"
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

func (c Client) GetVoiceSettings(ctx context.Context, voiceId string) (VoiceSettingsResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/settings"
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

func (c Client) GetVoice(ctx context.Context, voiceId string) (VoiceResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}"
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

func (c Client) GetVoices(ctx context.Context) (GetVoicesResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices"
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
