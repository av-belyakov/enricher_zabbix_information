package sseserver

import "github.com/av-belyakov/enricher_zabbix_information/interfaces"

func New(logger interfaces.Logger) *SSEServer {
	return &SSEServer{
		logger: logger,
		customerRegistration: SSEServerCustomerRegistration{
			clients: make(map[*Client]bool),
		},
	}
}

// WithPort устанавливает порт для взаимодействия с модулем
func WithPort(v int) sseServerOptions {
	return func(s *SSEServer) error {
		s.settings.port = v

		return nil
	}
}

// WithHost устанавливает хост для взаимодействия с модулем
func WithHost(v string) sseServerOptions {
	return func(s *SSEServer) error {
		s.settings.host = v

		return nil
	}
}
