package websocketserver

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

// ServerWs обработчик websocket
func ServeWs(logger interfaces.Logger, h *Hub, w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
		//ReadBufferSize:    1024,
		//WriteBufferSize:   100000000,
		HandshakeTimeout: (time.Duration(1) * time.Second),
	}

	// инициализация websocket соединения
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Send("error", wrappers.WrapperError(err).Error())

		return
	}

	client := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		logger: logger,
	}

	// регистрация нового клиента
	h.chRegister <- client

	logger.Send("info", "connection established, new client registration")

	// горутины для чтения и записи
	go client.writePump()
	go client.readPump(h)
}
