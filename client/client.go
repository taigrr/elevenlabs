package client

const apiEndpoint = "https://api.elevenlabs.io"

var ErrUnauthorized error

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
