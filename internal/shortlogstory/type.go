package shortlogstory

import "sync"

type ShortLogStory struct {
	mux   sync.RWMutex
	story []LogInformation
	size  int // ограничение на размер хранилища
}

type LogInformation struct {
	Date        string `json:"timestamp"`
	Type        string `json:"level"`
	Description string `json:"message"`
}
