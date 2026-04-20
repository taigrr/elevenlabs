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

func (c Client) HistoryDelete(ctx context.Context, historyItemID string) (bool, error) {
	url := fmt.Sprintf(c.endpoint+"/v1/history/%s", historyItemID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	res, err := c.httpClient.Do(req)

	switch res.StatusCode {
	case 401:
		return false, ErrUnauthorized
	case 200:
		if err != nil {
			return false, err
		}
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

func (c Client) HistoryDownloadZip(ctx context.Context, id1, id2 string, additionalIDs ...string) (io.ReadCloser, error) {
	url := c.endpoint + "/v1/history/download"
	downloads := append(additionalIDs, id1, id2)
	toDownload := types.HistoryPost{
		HistoryItemIds: downloads,
	}
	body, _ := json.Marshal(toDownload)
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "archive/zip")
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	res, err := c.httpClient.Do(req)

	switch res.StatusCode {
	case 401:
		return nil, ErrUnauthorized
	case 200:
		if err != nil {
			return nil, err
		}
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

func (c Client) HistoryDownloadAudio(ctx context.Context, ID string) (io.ReadCloser, error) {
	url := fmt.Sprintf(c.endpoint+"/v1/history/%s/audio", ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

func (c Client) GetHistoryItemList(ctx context.Context) ([]types.HistoryItemList, error) {
	url := c.endpoint + "/v1/history"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []types.HistoryItemList{}, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return []types.HistoryItemList{}, err
	}

	switch res.StatusCode {
	case 401:
		return []types.HistoryItemList{}, ErrUnauthorized
	case 200:
		var history types.GetHistoryResponse
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&history)
		if jerr != nil {
			return []types.HistoryItemList{}, jerr
		}
		return history.History, err
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
		return []types.HistoryItemList{}, err
	}
}

func (c Client) GetHistoryIDs(ctx context.Context, voiceIDs ...string) ([]string, error) {
	items, err := c.GetHistoryItemList(ctx)
	if err != nil {
		return []string{}, err
	}
	ids := []string{}
	voiceMap := make(map[string]struct{})
	for _, v := range voiceIDs {
		voiceMap[v] = struct{}{}
	}
	for _, i := range items {
		if len(voiceIDs) == 0 {
			ids = append(ids, i.HistoryItemID)
		} else {
			if _, ok := voiceMap[i.VoiceID]; ok {
				ids = append(ids, i.HistoryItemID)
			}
		}
	}
	return ids, err
}
