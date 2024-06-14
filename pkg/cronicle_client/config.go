package cronicle_client

import (
	"errors"
	"time"
)

type Config struct {
	BaseUrl       string
	APIKey        string
	Timeout       time.Duration
	RetryAttempts int
}

func (c *Config) Validate() error {
	if c.BaseUrl == "" {
		return errors.New("BaseUrl is required")
	}

	if c.APIKey == "" {
		return errors.New("APIKey is required")
	}

	if c.Timeout <= 0 {
		return errors.New("Timeout must be greater than 0")
	}

	if c.RetryAttempts < 0 {
		return errors.New("RetryAttempts must be greater than or equal to 0")
	}
	return nil
}
