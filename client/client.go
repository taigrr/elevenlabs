package client

import (
	"errors"
)

const apiEndpoint = "https://api.elevenlabs.io"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrUnspecified  = errors.New("unspecified error")
)

type Client struct {
	apiKey   string
	endpoint string
}

func New(apiKey string) Client {
	return Client{
		apiKey:   apiKey,
		endpoint: apiEndpoint,
	}
}

func (c Client) WithEndpoint(endpoint string) Client {
	c.endpoint = endpoint
	return c
}
