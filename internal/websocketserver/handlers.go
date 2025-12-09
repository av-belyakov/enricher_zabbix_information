package websocketserver

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func serveWs(h *Hub, w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
		//ReadBufferSize:    1024,
		//WriteBufferSize:   100000000,
		HandshakeTimeout: (time.Duration(1) * time.Second),
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	// Запускаем горутины для чтения и записи
	go client.writePump()
	go client.readPump(h)
}
