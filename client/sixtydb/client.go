// Package sixtydb is a thin Go client for 60db's TTS + STT + LLM APIs,
// shaped after the parent github.com/taigrr/elevenlabs/client package so
// callers can swap providers with a minimal import change.
//
// Reference: https://docs.60db.ai
package sixtydb

import (
	"errors"
	"net/http"
)

const apiEndpoint = "https://api.60db.ai"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrUnspecified  = errors.New("unspecified error")
)

// Client follows the same value-semantics builder pattern as the
// ElevenLabs Client in the parent package — With* methods return a copy
// so the Client is safe to share across goroutines without locks.
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

func (c Client) WithHTTPClient(hc *http.Client) Client {
	c.httpClient = hc
	return c
}
