package client

type Client struct {
	apiKey string
}

func New(apiKey string) Client {
	return Client{apiKey: apiKey}
}
