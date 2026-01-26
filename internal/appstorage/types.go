package appstorage

import (
	"net/netip"
	"sync"
	"sync/atomic"
	"time"
)

// SharedAppStorage общее хранилище приложения
type SharedAppStorage struct {
	statistics StatisticsApp
	logs       LogsApp
}

type Options func(*SharedAppStorage) error

// Statistics статистическая информация приложения
type StatisticsApp struct {
	mutex                     sync.RWMutex
	data                      []HostDetailedInformation
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
	Ips           []netip.Addr `json:"ips"`             // список ip адресов
	SensorsId     []string     `json:"sensor_id"`       // id обслуживающего сенсора
	NetboxHostsId []int        `json:"netbox_hosts_id"` // id хоста в netbox
	OriginalHost  string       `json:"original_host"`   // исходное наименование хоста
	DomainName    string       `json:"domain_name"`     // доменное имя
	Error         error        `json:"error"`           // ошибка
	HostId        int          `json:"host_id"`         // id хоста
	IsActive      bool         `json:"is_active"`       // флаг активный ли хост
	IsProcessed   bool         `json:"is_processed"`    // флаг обработан ли хост
}

// LogsApp логи приложения
type LogsApp struct {
	mutex sync.RWMutex
	story []LogInformation
	size  int // ограничение на размер хранилища
}

// LogInformation информация по логам
type LogInformation struct {
	Date        string `json:"timestamp"`
	Type        string `json:"level"`
	Description string `json:"message"`
}
