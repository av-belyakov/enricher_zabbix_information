package storage

import (
	"net/netip"
	"sync"
	"sync/atomic"
)

// ShortTermStorage хранилище в памяти
type ShortTermStorage struct {
	data        []HostDetailedInformation
	mutex       sync.RWMutex
	isExecution atomic.Bool
}

// HostDetailedInformation детальная информация о хосте
type HostDetailedInformation struct {
	Ips          []netip.Addr // список ip адресов
	OriginalHost string       // исходное наименование хоста
	DomainName   string       // доменное имя
	Error        error        // ошибка
	HostId       int          // id хоста
}
