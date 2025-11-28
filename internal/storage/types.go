package storage

import (
	"net/netip"
	"sync"
)

// ShortTermStorage хранилище в памяти
type ShortTermStorage struct {
	mutex sync.RWMutex
	data  []HostDetailedInformation
}

// HostDetailedInformation детальная информация о хосте
type HostDetailedInformation struct {
	Ips          []netip.Addr // список ip адресов
	OriginalHost string       // исходное наименование хоста
	DomainName   string       // доменное имя
	Error        error        // ошибка
	HostId       int          // id хоста
}
