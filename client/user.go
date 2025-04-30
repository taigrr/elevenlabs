package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/taigrr/elevenlabs/client/types"
)

func (c Client) GetUserInfo(ctx context.Context) (types.UserResponseModel, error) {
	url := c.endpoint + "/v1/user"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return types.UserResponseModel{}, err
	}
	req.Header.Set("xi-api-key", c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs")
	req.Header.Set("accept", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return types.UserResponseModel{}, err
	}
	switch res.StatusCode {
	case 401:
		return types.UserResponseModel{}, ErrUnauthorized
	case 200:
		var user types.UserResponseModel
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&user)
		if jerr != nil {
			return types.UserResponseModel{}, jerr
		}
		return user, err
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
		return types.UserResponseModel{}, err
	}
}

func (c Client) GetSubscriptionInfo(ctx context.Context) (types.Subscription, error) {
	info, err := c.GetUserInfo(ctx)
	if err != nil {
		return types.Subscription{}, err
	}
	return info.Subscription, err
}
