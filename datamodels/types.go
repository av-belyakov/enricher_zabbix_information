package datamodels

import (
	"net/netip"
)

// HostDetailedInformation детальная информация о хосте
type HostDetailedInformation struct {
	Ips          []netip.Addr // список ip адресов
	OriginalHost string       // исходное наименование хоста
	DomainName   string       // доменное имя
	Error        error        // ошибка
	HostId       int          // id хоста
	IsProcessed  bool         // флаг обработан ли хост
}

// TemplBasePage тип для templ, базовая страница
type TemplBasePage struct {
	Title        string
	AppName      string
	AppVersion   string
	AppShortInfo string
	MenuLinks    []struct {
		Name string
		Link string
		Icon string
	}
}

// TemplTaskCompletionsStatistics тип для templ, статистика выполненной задачи
type TemplTaskCompletionsStatistics struct {
	Hosts []struct {
		Name  string `json:"name"`
		Error string `json:"error"`
	} `json:"hosts"`
	DataStart             string `json:"data_start"`
	DataEnd               string `json:"data_end"`
	DiffTime              string `json:"diff_time"`
	ExecutionStatus       string `json:"execution_status"`
	CountHosts            int    `json:"count_hosts"`
	CountHostsError       int    `json:"count_hosts_error"`
	CountHostsIsProcessed int    `json:"count_hosts_is_processed"`
}
