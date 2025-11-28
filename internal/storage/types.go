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
	Ip           []netip.Addr // список ip адресов
	Errors       error        // ошибка
	OriginalHost string       // исходное наименование хоста
	DomainName   string       // доменное имя
	HostId       int          // id хоста
}
