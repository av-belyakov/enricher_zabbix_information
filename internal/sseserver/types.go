package sseserver

import (
	"sync"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// Client параметры клиента
type Client struct {
	messages chan string
}

// SSEServer SSE сервер
type SSEServer struct {
	customerRegistration SSEServerCustomerRegistration
	settings             SSEServerSettings
	logger               interfaces.Logger
}

// SSEServer настройки SSE сервера
type SSEServerSettings struct {
	host string
	port int
}

// SSEServerCustomerRegistration регистрация клиентов
type SSEServerCustomerRegistration struct {
	clients map[*Client]bool
	mutex   sync.RWMutex
}

type sseServerOptions func(*SSEServer) error
