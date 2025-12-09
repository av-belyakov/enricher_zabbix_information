package sseserver

import (
	"net/http"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

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

func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		next(w, r)
	}
}
