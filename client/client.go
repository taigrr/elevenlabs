package client

import (
	"errors"
	"io"
	"net/http"
)

const (
	apiEndpoint = "https://api.elevenlabs.io"
	userAgent   = "github.com/taigrr/elevenlabs"

	headerAccept      = "accept"
	headerContentType = "Content-Type"
	headerUserAgent   = "User-Agent"
	headerAPIKey      = "xi-api-key"

	contentTypeJSON = "application/json"
	contentTypeMP3  = "audio/mpeg"
	contentTypeZip  = "archive/zip"
)

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
		httpClient: defaultHTTPClient(),
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

// WithHTTPClient allows users to provide their own http.Client.
// Passing nil resets the client to the default HTTP client.
func (c Client) WithHTTPClient(hc *http.Client) Client {
	if hc == nil {
		hc = defaultHTTPClient()
	}
	c.httpClient = hc
	return c
}

func defaultHTTPClient() *http.Client {
	return &http.Client{}
}
