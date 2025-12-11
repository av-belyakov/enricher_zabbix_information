package websocketserver

// New конструктор нового хаба
func New() *Hub {
	return &Hub{
		chBroadcast:    make(chan []byte),
		chIncomingData: make(chan []byte),
		chRegister:     make(chan *Client),
		chUnregister:   make(chan *Client),
		clients:        make(map[*Client]bool),
	}
}
