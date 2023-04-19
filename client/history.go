package client

import (
	"bufio"
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
		defer res.Body.Close()
		io.Copy(w, res.Body)
		return nil
	default:
		return errors.Join(err, ErrUnspecified)
	}
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
		return []byte{}, err
	case 401:
		return []byte{}, ErrUnauthorized
	case 200:
		if err != nil {
			return []byte{}, err
		}
		b := bytes.Buffer{}
		w := bufio.NewWriter(&b)

		defer res.Body.Close()
		io.Copy(w, res.Body)
		return b.Bytes(), nil
	default:
		return []byte{}, errors.Join(err, ErrUnspecified)
	}
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
	req.Header.Set("accept", "audio/mpeg")
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
		defer res.Body.Close()
		io.Copy(w, res.Body)
		return nil
	default:
		return errors.Join(err, ErrUnspecified)
	}
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
	req.Header.Set("accept", "audio/mpeg")
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
		return []byte{}, err
	case 401:
		return []byte{}, ErrUnauthorized
	case 200:
		if err != nil {
			return []byte{}, err
		}
		b := bytes.Buffer{}
		w := bufio.NewWriter(&b)

		defer res.Body.Close()
		io.Copy(w, res.Body)
		return b.Bytes(), nil
	default:
		return []byte{}, errors.Join(err, ErrUnspecified)
	}
}

func (c Client) GetHistoryItemList(ctx context.Context) ([]types.HistoryItemList, error) {
	url := c.endpoint + "/v1/history"
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []types.HistoryItemList{}, err
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
		return []types.HistoryItemList{}, err
	case 401:
		return []types.HistoryItemList{}, ErrUnauthorized
	case 200:
		if err != nil {
			return []types.HistoryItemList{}, err
		}

		var history types.GetHistoryResponse
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&history)
		if jerr != nil {
			return []types.HistoryItemList{}, jerr
		}
		return history.History, err

	default:
		return []types.HistoryItemList{}, errors.Join(err, ErrUnspecified)
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
