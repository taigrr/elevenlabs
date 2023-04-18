package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/taigrr/elevenlabs/client/types"
)

var ErrUnauthorized error

func (c Client) HistoryDelete(ctx context.Context, historyItemId string) (bool, error) {
	// create path and map variables
	url := fmt.Sprintf(apiEndpoint+"/v1/history/%s", historyItemId)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	res, err := client.Do(req)

	switch res.StatusCode {
	case 422:
		return false, err
	case 401:
		return false, ErrUnauthorized
	case 200:
		if err != nil {
			return false, nil
		}
	}
	return true, nil
}

func (c Client) HistoryDownload(ctx context.Context, id string, additionalIDs ...string) ([]byte, error) {
	url := apiEndpoint + "/v1/history/download"
	downloads := append(additionalIDs, id)
	toDownload := types.HistoryPost{
		HistoryItemIds: downloads,
	}
	client := &http.Client{}
	body, _ := json.Marshal(toDownload)
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return []byte{}, err
	}
	if len(downloads) == 1 {
		req.Header.Set("accept", "audio/mpeg")
	} else {
		req.Header.Set("accept", "archive/zip")
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	res, err := client.Do(req)
}

func (c Client) DownloadAudioByID(ctx context.Context, ID string) ([]byte, error) {
	url := fmt.Sprintf(apiEndpoint+"/v1/history/%s/audio", ID)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
}

func (c Client) GetHistoryList(ctx context.Context) ([]string, error) {
	url := apiEndpoint + "/v1/history"
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []string{}, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
}
