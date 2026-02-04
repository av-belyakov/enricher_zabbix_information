package datamodels

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
	ProcessedHosts []struct {
		SensorsId    string `json:"sensors_id"`    // список id обслуживающих сенсоров
		OriginalHost string `json:"original_host"` // исходное наименование хоста
		HostId       int    `json:"host_id"`       // id хоста
		Ips          string `json:"ips"`           // список ip адресов
	} `json:"processed_hosts"`
	DataStart                   string `json:"data_start"`
	DataEnd                     string `json:"data_end"`
	DiffTime                    string `json:"diff_time"`
	ExecutionStatus             string `json:"execution_status"`
	CountHostsError             int    `json:"count_hosts_error"`              // количество доменных имён обработанных с ошибкой
	CountFoundIpToPrefix        int    `json:"count_found_ip_to_prefix"`       // количество ip адресов совпавших с префиксами Netbox
	CountZabbixHostsGroup       int    `json:"count_zabbix_hosts_group"`       // количество групп хостов в Zabbix
	CountZabbixHosts            int    `json:"count_zabbix_hosts"`             // общее количество хостов в Zabbix
	CountMonitoringHostsGroup   int    `json:"count_monitoring_hosts_group"`   // количество групп хостов по которым осуществляется мониторинг
	CountMonitoringHosts        int    `json:"count_monitoring_hosts"`         // количество хостов по которым осуществляется мониторинг
	CountNetboxPrefixes         int    `json:"count_netbox_prefixes"`          // количество найденных префиксов в Netbox
	CountNetboxPrefixesReceived int    `json:"count_netbox_prefixes_received"` // количество полученых из Netbox префиксов
	CountNetboxPrefixesMatches  int    `json:"count_netbox_prefixes_matches"`  // количество префиксов в которых найдено совпадение
	CountUpdatedZabbixHosts     int    `json:"count_updated_zabbix_hosts"`     // количество обновленных хостов в Zabbix
}
