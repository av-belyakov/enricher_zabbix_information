package dnsresolver

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// New конструктор
func New(storage interfaces.StorageDNSResolver, opts ...Options) (*Settings, error) {
	settings := &Settings{
		storage: storage,
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, address)
			},
		},
		timeout: time.Second * 15,
	}

	for _, opt := range opts {
		if err := opt(settings); err != nil {
			return settings, err
		}
	}

	return settings, nil
}

// WithTimeout время ожидания выполнения запроса
func WithTimeout(timeout int) Options {
	return func(s *Settings) error {
		if timeout == 0 {
			return errors.New("timeout cannot be zero")
		}

		return nil
	}
}
