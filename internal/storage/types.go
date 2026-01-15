package storage

import (
	"net/netip"
	"sync"
	"sync/atomic"
	"time"
)

// ShortTermStorage хранилище в памяти
type ShortTermStorage struct {
	data               []HostDetailedInformation
	mutex              sync.RWMutex
	startDateExecution time.Time
	endDateExecution   time.Time
	isExecution        atomic.Bool
}

// HostDetailedInformation детальная информация о хосте
type HostDetailedInformation struct {
	Ips           []netip.Addr // список ip адресов
	SensorsId     []string     // id обслуживающего сенсора
	NetboxHostsId []int        // id хоста в netbox
	OriginalHost  string       // исходное наименование хоста
	DomainName    string       // доменное имя
	Error         error        // ошибка
	HostId        int          // id хоста
	IsActive      bool         // флаг активный ли хост
	IsProcessed   bool         // флаг обработан ли хост
}
