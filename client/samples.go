package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/taigrr/elevenlabs/client/types"
)

func (c Client) DeleteVoiceSample(ctx context.Context, voiceID, sampleID string) (bool, error) {
	url := fmt.Sprintf(c.endpoint+"/v1/voices/%s/samples/%s", voiceID, sampleID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set(headerAccept, contentTypeJSON)
	req.Header.Set(headerAPIKey, c.apiKey)
	req.Header.Set(headerUserAgent, userAgent)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	switch res.StatusCode {
	case 401:
		res.Body.Close()
		return false, ErrUnauthorized
	case 200:
		drainClose(res.Body)
		return true, nil
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
		return false, err
	}
}

func (c Client) DownloadVoiceSample(ctx context.Context, voiceID, sampleID string) (io.ReadCloser, error) {
	url := fmt.Sprintf(c.endpoint+"/v1/voices/%s/samples/%s/audio", voiceID, sampleID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerAPIKey, c.apiKey)
	req.Header.Set(headerUserAgent, userAgent)
	req.Header.Set(headerAccept, contentTypeMP3)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch res.StatusCode {
	case 401:
		res.Body.Close()
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
