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
	OriginalHost string       // исходное наименование хоста
	DomainName   string       // доменное имя
	Errors       error        // ошибка
	HostId       int          // id хоста
}
