package logginghandler

import (
	"slices"
	"sync"
)

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

func NewShortLogStory(size int) *ShortLogStory {
	return &ShortLogStory{
		story: make([]LogInformation, 0),
		size:  size,
	}
}

// Add добавить информацию по логам
func (ls *ShortLogStory) Add(v LogInformation) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	if len(ls.story) == ls.size {
		ls.story = append(ls.story[1:], v)

		return
	}

	ls.story = append(ls.story, v)
}

// Get получить информицию по логам
func (ls *ShortLogStory) Get() []LogInformation {
	ls.mux.RLock()
	defer ls.mux.RUnlock()

	newList := make([]LogInformation, len(ls.story))
	copy(newList, ls.story)

	slices.Reverse(newList)

	return newList
}
