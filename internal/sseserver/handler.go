package sseserver

import (
	"fmt"
	"net/http"
	"time"
)

// HandleSSE обработчик SSE соединений
func (s *SSEServer) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем заголовки SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Создаем клиента
	client := &Client{
		messages: make(chan string, 10),
	}

	fmt.Println("func 'SSEServer.HandleSSE', Регистрируем клиента")
	// Регистрируем клиента
	s.AddClient(client)
	defer s.RemoveClient(client)

	fmt.Println("func 'SSEServer.HandleSSE', Отправляем приветственное сообщение")
	// Отправляем приветственное сообщение
	fmt.Fprintf(w, "data: Connected to server")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Таймер для heartbeat
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			// Обработка закрытия соединения
			fmt.Println("func 'SSEServer.HandleSSE', SSE соединение закрыто")

			return

		case msg := <-client.messages:
			fmt.Println("func 'SSEServer.HandleSSE', SSE message")

			fmt.Fprintf(w, "data: %s\n\n", msg)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

		case <-ticker.C:
			fmt.Println("func 'SSEServer.HandleSSE', SSE heartbeat")

			fmt.Fprintf(w, ": heartbeat\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}
