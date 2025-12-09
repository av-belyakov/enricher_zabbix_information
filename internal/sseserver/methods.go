package sseserver

import (
	"fmt"
)

// AddClient добавляет нового клиента
func (s *SSEServer) AddClient(client *Client) {
	s.customerRegistration.mutex.Lock()
	defer s.customerRegistration.mutex.Unlock()

	s.customerRegistration.clients[client] = true

	s.logger.Send("info", fmt.Sprintf("Добавлен клиент. Всего: %d", len(s.customerRegistration.clients)))
	//log.Printf("Добавлен клиент. Всего: %d", len(s.customerRegistration.clients))
}

// RemoveClient удаляет клиента
func (s *SSEServer) RemoveClient(client *Client) {
	s.customerRegistration.mutex.Lock()
	defer s.customerRegistration.mutex.Unlock()

	delete(s.customerRegistration.clients, client)
	close(client.messages)

	s.logger.Send("info", fmt.Sprintf("Удален клиент. Всего: %d", len(s.customerRegistration.clients)))
	//log.Printf("Удален клиент. Всего: %d", len(s.customerRegistration.clients))
}

// Broadcast рассылает сообщение всем клиентам
func (s *SSEServer) Broadcast(message string) {
	s.customerRegistration.mutex.RLock()
	defer s.customerRegistration.mutex.RUnlock()

	for client := range s.customerRegistration.clients {
		select {
		case client.messages <- message:
		default:
			// Пропускаем заблокированных клиентов
		}
	}
}
