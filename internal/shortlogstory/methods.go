package shortlogstory

import "slices"

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
