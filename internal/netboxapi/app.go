package netboxapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

func New(token string, opts ...Options) (*Client, error) {
	if token == "" {
		return nil, errors.New("the token must not be empty")
	}

	client := &Client{
		settings: Settings{
			host:    "localhost",
			port:    8005,
			timeout: 15,
			token:   token,
		},
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return client, err
		}
	}

	return client, nil
}

// WithHost имя или ip адрес хоста API
func WithHost(v string) Options {
	return func(c *Client) error {
		if v == "" {
			return errors.New("the value of 'host' cannot be empty")
		}

		c.settings.host = v

		return nil
	}
}

// WithPort сетевой порт API
func WithPort(v int) Options {
	return func(c *Client) error {
		if v <= 0 || v > 65535 {
			return errors.New("an incorrect network port value was received")
		}

		c.settings.port = v

		return nil
	}
}

// WithTimeout время ожидания ответа от API
func WithTimeout(v int) Options {
	return func(c *Client) error {
		timeoutMin, timeoutMax := 1, 6_000

		if v <= timeoutMin || timeoutMax > 6_000 {
			return fmt.Errorf("The response waiting time should be in the range from %d to %d seconds", timeoutMin, timeoutMax)
		}

		c.settings.timeout = v

		return nil
	}
}
