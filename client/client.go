package client

const apiEndpoint = "https://api.elevenlabs.io"

type Client struct {
	apiKey string
}

func New(apiKey string) Client {
	return Client{apiKey: apiKey}
}
