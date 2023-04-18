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

func (c Client) HistoryDelete(ctx context.Context, historyItemId string) (bool, error) {
	url := fmt.Sprintf(c.endpoint+"/v1/history/%s", historyItemId)

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
		ve := types.ValidationError{}
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&ve)
		if jerr != nil {
			err = errors.Join(err, jerr)
		} else {
			err = errors.Join(err, ve)
		}
		return false, err
	case 401:
		return false, ErrUnauthorized
	case 200:
		if err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, errors.Join(err, ErrUnspecified)
	}
	return true, nil
}

func (c Client) HistoryDownloadZipWriter(ctx context.Context, w io.Writer, id1, id2 string, additionalIDs ...string) error {
	url := c.endpoint + "/v1/history/download"
	downloads := append(additionalIDs, id1, id2)
	toDownload := types.HistoryPost{
		HistoryItemIds: downloads,
	}
	client := &http.Client{}
	body, _ := json.Marshal(toDownload)
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("accept", "archive/zip")
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	res, err := client.Do(req)
	io.Copy(w, req.Response.Body)
}

func (c Client) HistoryDownloadZip(ctx context.Context, id1, id2 string, additionalIDs ...string) ([]byte, error) {
	url := c.endpoint + "/v1/history/download"
	downloads := append(additionalIDs, id1, id2)
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
	req.Header.Set("accept", "archive/zip")
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	res, err := client.Do(req)
}

func (c Client) HistoryDownloadAudioWriter(ctx context.Context, w io.Writer, ID string) error {
	url := fmt.Sprintf(c.endpoint+"/v1/history/%s/audio", ID)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	res, err := client.Do(req)
	io.Copy(w, req.Response.Body)
}

func (c Client) HistoryDownloadAudio(ctx context.Context, ID string) ([]byte, error) {
	url := fmt.Sprintf(c.endpoint+"/v1/history/%s/audio", ID)
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
}

func (c Client) GetHistoryList(ctx context.Context) ([]string, error) {
	url := c.endpoint + "/v1/history"
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []string{}, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
}
