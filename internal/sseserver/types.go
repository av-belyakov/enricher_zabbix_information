package sseserver

import "sync"

type Client struct {
	messages chan string
}

type SSEServer struct {
	clients map[*Client]bool
	mutex   sync.RWMutex
}
