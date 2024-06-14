package cronicle_client

import (
	"net/http"
)

// Client is a struct that holds the base URL and API key
type Client struct {
	httpClient *http.Client
	config     Config
}

// NewClient is a function that creates a new Client
func NewClient(config Config) *Client {

	err := config.Validate()
	if err != nil {
		panic(err)
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}
