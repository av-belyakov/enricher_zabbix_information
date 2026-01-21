package storage

import (
	"net/netip"
	"sync"
	"sync/atomic"
	"time"
)

// ShortTermStorage хранилище в памяти
type ShortTermStorage struct {
	data                      []HostDetailedInformation
	mutex                     sync.RWMutex
	startDateExecution        time.Time
	endDateExecution          time.Time
	countZabbixHostsGroup     atomic.Int32 // количество групп хостов в Zabbix
	countZabbixHosts          atomic.Int32 // общее количество хостов в Zabbix
	countMonitoringHostsGroup atomic.Int32 // количество групп хостов по которым осуществляется мониторинг
	countMonitoringHosts      atomic.Int32 // количество хостов по которым осуществляется мониторинг
	countNetboxPrefixes       atomic.Int32 // количество найденных префиксов в Netbox
	countUpdatedZabbixHosts   atomic.Int32 // количество обновленных хостов в Zabbix
	isExecution               atomic.Bool
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
