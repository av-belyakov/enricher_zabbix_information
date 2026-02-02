package supportingfunctions

import (
	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

const Zero_Pattern = "00:00:00 0001-01-01"

// CreateTaskStatistics создание статистики по выполняемой задаче
func CreateTaskStatistics(storage interfaces.StorageInformation) datamodels.TemplTaskCompletionsStatistics {
	taskStatus := "завершена"
	if storage.GetStatusProcessRunning() {
		taskStatus = "выполняется"
	}

	start, end := storage.GetDateExecution()
	dateStart := start.Format("15:04:05 2006-01-02")
	dateEnd := end.Format("15:04:05 2006-01-02")
	diffTime := "-"
	if dateEnd != Zero_Pattern {
		diffTime = end.Sub(start).String()
	}

	listHostsError := []struct {
		Name  string `json:"name"`
		Error string `json:"error"`
	}{}

	listProcessedHosts := []struct {
		SensorsId    []string `json:"sensor_id"`     // id обслуживающего сенсора
		OriginalHost string   `json:"original_host"` // исходное наименование хоста
		HostId       int      `json:"host_id"`       // id хоста
	}{}

	hosts := storage.GetList()
	var countFoundIpToPrefix int
	for _, v := range hosts {
		if v.IsProcessed {
			countFoundIpToPrefix++

			listProcessedHosts = append(listProcessedHosts, struct {
				SensorsId    []string `json:"sensor_id"`
				OriginalHost string   `json:"original_host"`
				HostId       int      `json:"host_id"`
			}{
				SensorsId:    v.SensorsId,
				OriginalHost: v.OriginalHost,
				HostId:       v.HostId,
			})
		}

		if v.Error == nil {
			continue
		}

		listHostsError = append(
			listHostsError,
			struct {
				Name  string `json:"name"`
				Error string `json:"error"`
			}{
				Name:  v.OriginalHost,
				Error: v.Error.Error(),
			})
	}

	dEnd := "-"
	if dateEnd != Zero_Pattern {
		dEnd = dateEnd
	}

	return datamodels.TemplTaskCompletionsStatistics{
		DataStart:                    dateStart,
		DataEnd:                      dEnd,
		DiffTime:                     diffTime,
		ExecutionStatus:              taskStatus,
		CountHostsError:              len(listHostsError),
		CountFoundIpToPrefix:         countFoundIpToPrefix,
		CountZabbixHostsGroup:        int(storage.GetCountZabbixHostsGroup()),
		CountZabbixHosts:             int(storage.GetCountZabbixHosts()),
		CountMonitoringHostsGroup:    int(storage.GetCountMonitoringHostsGroup()),
		CountMonitoringHosts:         int(storage.GetCountMonitoringHosts()),
		CountNetboxPrefixes:          int(storage.GetCountNetboxPrefixes()),
		CountNetboxPrefixesReceived:  int(storage.GetCountNetboxPrefixesReceived()),
		CountNetboxPrefixesProcessed: int(storage.GetCountNetboxPrefixesProcessed()),
		CountUpdatedZabbixHosts:      int(storage.GetCountUpdatedZabbixHosts()),
		Hosts:                        listHostsError,
		ProcessedHosts:               listProcessedHosts,
	}
}
