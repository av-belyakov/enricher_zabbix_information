package supportingfunctions

import (
	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// CreateTaskStatistics создание статистики по выполняемой задаче
func CreateTaskStatistics(storage interfaces.StorageInformation) datamodels.TemplTaskCompletionsStatistics {
	taskStatus := "завершена"
	if storage.GetStatusProcessRunning() {
		taskStatus = "выполняется"
	}

	start, end := storage.GetDateExecution()
	dateStart := start.Format("15:04:05 2006-01-02")
	dateEnd := "-"
	diffTime := "-"
	if end.String() != "00:00:00 0001-01-01" {
		dateEnd = end.Format("15:04:05 2006-01-02")
		diffTime = end.Sub(start).String()
	}

	listHostsError := []struct {
		Name  string `json:"name"`
		Error error  `json:"error"`
	}{}
	hosts := storage.GetList()
	var countHostIsProcessed int
	for _, v := range hosts {
		if v.IsProcessed {
			countHostIsProcessed++
		}

		if v.Error == nil {
			continue
		}

		listHostsError = append(
			listHostsError,
			struct {
				Name  string `json:"name"`
				Error error  `json:"error"`
			}{
				Name:  v.OriginalHost,
				Error: v.Error,
			})
	}

	return datamodels.TemplTaskCompletionsStatistics{
		DataStart:             dateStart,
		DataEnd:               dateEnd,
		DiffTime:              diffTime,
		ExecutionStatus:       taskStatus,
		CountHosts:            len(hosts),
		CountHostsError:       len(listHostsError),
		CountHostsIsProcessed: countHostIsProcessed,
		Hosts:                 listHostsError,
	}
}
