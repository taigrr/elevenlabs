package client

const apiEndpoint = "https://api.elevenlabs.io"

var ErrUnauthorized error

type Client struct {
	apiKey string
}

func New(apiKey string) Client {
	return Client{apiKey: apiKey}
}
