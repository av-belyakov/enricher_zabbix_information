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
		Name  string
		Error error
	}
	DataStart       string
	DataEnd         string
	DiffTime        string
	ExecutionStatus string
	CountHosts      int
	CountHostsError int
}
