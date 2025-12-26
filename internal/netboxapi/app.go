package netboxapi

import (
	"cmp"
	"errors"
	"net/http"
	"time"
)

func New(host string, port int, token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("the token must not be empty")
	}

	if host == "" {
		return nil, errors.New("the host must not be empty")
	}

	return &Client{
		settings: Settings{
			host:  host,
			port:  cmp.Or(port, 8005),
			token: token,
		},
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}, nil
}
