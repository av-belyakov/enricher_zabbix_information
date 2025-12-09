package sseserver

import (
	"fmt"
	"net/http"
	"time"
)

// HandleSSE обработчик запросов SSE
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

	// Регистрируем клиента
	s.AddClient(client)
	defer s.RemoveClient(client)

	// Отправляем приветственное сообщение
	fmt.Fprintf(w, "data: Connected to server\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Таймер для heartbeat
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Обработка закрытия соединения
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-client.messages:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}
