package sseserver

import "log"

// AddClient добавляет нового клиента
func (s *SSEServer) AddClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.clients[client] = true

	log.Printf("Добавлен клиент. Всего: %d", len(s.clients))
}

// RemoveClient удаляет клиента
func (s *SSEServer) RemoveClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.clients, client)
	close(client.messages)

	log.Printf("Удален клиент. Всего: %d", len(s.clients))
}

// Broadcast рассылает сообщение всем клиентам
func (s *SSEServer) Broadcast(message string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for client := range s.clients {
		select {
		case client.messages <- message:
		default:
			// Пропускаем заблокированных клиентов
		}
	}
}
