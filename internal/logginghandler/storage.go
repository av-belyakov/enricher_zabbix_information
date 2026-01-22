package logginghandler

import "sync"

type ShortLogStory struct {
	mux   sync.RWMutex
	story []LogInformation
	size  int // ограничение на размер хранилища
}

type LogInformation struct {
	Type, Description string
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

	ls.story = append(ls.story, v)
}
