package websocketserver

import "github.com/gorilla/websocket"

// Клиент представляет подключение
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// Хаб управляет всеми клиентами
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}
