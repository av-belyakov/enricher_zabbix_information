package sseserver

/*

Этот сервер надо внимательно посмотреть. Проверить логику.

*/

func New() *SSEServer {
	return &SSEServer{
		clients: make(map[*Client]bool),
	}
}
