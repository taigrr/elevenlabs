package client

import (
	"errors"
	"io"
	"net/http"
)

const apiEndpoint = "https://api.elevenlabs.io"

// drainClose drains and closes a response body so the underlying connection
// can be returned to the keep-alive pool. Use it on success paths that do not
// otherwise read the body.
func drainClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}

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
