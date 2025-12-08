package apiserver

import (
	"net/http"

	"github.com/av-belyakov/enricher_zabbix_information/components"
	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
)

// RouteTaskInformation информация о выполненной задаче
func (is *InformationServer) RouteTaskInformation(w http.ResponseWriter, r *http.Request) {
	taskStatus := "завершена"
	if is.storage.GetStatusProcessRunning() {
		taskStatus = "выполняется"
	}

	start, end := is.storage.GetDateExecution()
	dateStart := start.Format("15:04:05 2006-01-02")
	dateEnd := "-"
	diffTime := "-"
	if end.String() != "00:00:00 0001-01-01" {
		dateEnd = end.Format("15:04:05 2006-01-02")
		diffTime = end.Sub(start).String()
	}

	listHostsError := []struct {
		Name  string
		Error error
	}{}
	hosts := is.storage.GetList()
	for _, v := range hosts {
		if v.Error == nil {
			continue
		}

		listHostsError = append(
			listHostsError,
			struct {
				Name  string
				Error error
			}{
				Name:  v.OriginalHost,
				Error: v.Error,
			})
	}

	ttcs := datamodels.TemplTaskCompletionsStatistics{
		DataStart:       dateStart,
		DataEnd:         dateEnd,
		DiffTime:        diffTime,
		ExecutionStatus: taskStatus,
		CountHosts:      len(hosts),
		CountHostsError: len(listHostsError),
		Hosts:           listHostsError,
	}

	is.getBasePage(
		components.TemplateTaskCompletionStatistics(ttcs),
		components.BaseComponentScripts(),
	).Component.Render(r.Context(), w)
}
