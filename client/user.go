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
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return types.UserResponseModel{}, err
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
		return types.UserResponseModel{}, err
	case 401:
		return types.UserResponseModel{}, ErrUnauthorized
	case 200:
		if err != nil {
			return types.UserResponseModel{}, err
		}

		var user types.UserResponseModel
		defer res.Body.Close()
		jerr := json.NewDecoder(res.Body).Decode(&user)
		if jerr != nil {
			return types.UserResponseModel{}, jerr
		}
		return user, err

	default:
		return types.UserResponseModel{}, errors.Join(err, ErrUnspecified)
	}
}

func (c Client) GetSubscriptionInfo(ctx context.Context) (types.Subscription, error) {
	info, err := c.GetUserInfo(ctx)
	return info.Subscription, err
}
