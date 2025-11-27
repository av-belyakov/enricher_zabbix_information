package storage

import (
	"sync"
)

// ShortTermStorage хранилище в памяти
type ShortTermStorage struct {
	mutex sync.RWMutex
	Data  []HostDetailedInformation
}

// HostDetailedInformation детальная информация о хосте
type HostDetailedInformation struct {
	Ip           []string // список ip адресов
	Error        error    // ошибка
	OriginalHost string   // исходное наименование хоста
	DomainName   string   // доменное имя
	HostId       int      // id хоста
}
