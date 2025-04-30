package client

import (
	"errors"
	"net/http"
)

const apiEndpoint = "https://api.elevenlabs.io"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrUnspecified  = errors.New("unspecified error")
)

type Client struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

func New(apiKey string) Client {
	return Client{
		apiKey:     apiKey,
		endpoint:   apiEndpoint,
		httpClient: &http.Client{},
	}
}

func (c Client) WithEndpoint(endpoint string) Client {
	c.endpoint = endpoint
	return c
}

func (c Client) WithAPIKey(apiKey string) Client {
	c.apiKey = apiKey
	return c
}

// WithHTTPClient allows users to provide their own http.Client
func (c Client) WithHTTPClient(hc *http.Client) Client {
	c.httpClient = hc
	return c
}
