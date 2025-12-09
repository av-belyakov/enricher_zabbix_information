package websocketserver

import (
	"context"

	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
	"github.com/gorilla/websocket"
)

// Run запуск обработчика
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
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
}

// GetChanBroadcast канал для рассылки широковещательных сообщений
func (h *Hub) GetChanBroadcast() chan<- []byte {
	return h.broadcast
}

// SendBroadcast отправить широковещательное сообщение
func (h *Hub) SendBroadcast(b []byte) {
	h.broadcast <- b
}

func (c *Client) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Send("error", wrappers.WrapperError(err).Error())
			}

			break
		}

		// Отправляем полученное сообщение всем клиентам
		h.broadcast <- message
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		message, ok := <-c.send
		if !ok {
			// Канал закрыт
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})

			return
		}

		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			c.logger.Send("error", wrappers.WrapperError(err).Error())

			return
		}
	}
}
