package websocketserver

import (
	"github.com/gorilla/websocket"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// Client реализованное подключение
type Client struct {
	conn   *websocket.Conn
	logger interfaces.Logger
	send   chan []byte
}

// Hub управление всеми клиентами
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}
