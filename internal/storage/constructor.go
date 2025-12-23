package storage

// NewShortTermStorage конструктор кратковременного хранилища
func NewShortTermStorage() *ShortTermStorage {
	return &ShortTermStorage{
		data: []HostDetailedInformation{},
	}
}
