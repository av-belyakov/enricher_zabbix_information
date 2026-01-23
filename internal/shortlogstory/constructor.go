package shortlogstory

func NewShortLogStory(size int) *ShortLogStory {
	return &ShortLogStory{
		story: make([]LogInformation, 0),
		size:  size,
	}
}
