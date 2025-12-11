package websocketserver

import (
	"context"

	"github.com/gorilla/websocket"

	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

// Run запуск обработчика
func (h *Hub) Run(ctx context.Context) <-chan []byte {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case client := <-h.chRegister:
				h.clients[client] = true

			case client := <-h.chUnregister:
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
				}

			case message := <-h.chBroadcast:
				for client := range h.clients {
					select {
					case client.send <- message:

					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}()

	return h.chIncomingData
}

// GetChanBroadcast канал для рассылки широковещательных сообщений
//func (h *Hub) GetChanBroadcast() chan<- []byte {
//	return h.chBroadcast
//}

// SendBroadcast отправить широковещательное сообщение
func (h *Hub) SendBroadcast(b []byte) {
	h.chBroadcast <- b
}

func (c *Client) readPump(h *Hub) {
	defer func() {
		h.chUnregister <- c
		c.conn.Close()
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Send("error", wrappers.WrapperError(err).Error())
			}

			break
		}

		// Отправляем полученное сообщение на канал исходящих данных
		h.chIncomingData <- msg
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		msg, ok := <-c.send
		if !ok {
			// Канал закрыт
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})

			return
		}

		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			c.logger.Send("error", wrappers.WrapperError(err).Error())

			return
		}
	}
}
