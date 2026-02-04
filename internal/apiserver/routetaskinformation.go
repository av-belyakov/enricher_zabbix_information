package apiserver

import (
	"net/http"

	"github.com/av-belyakov/enricher_zabbix_information/components"
)

// RouteTaskInformation информация о выполненной задаче
func (is *InformationServer) RouteTaskInformation(w http.ResponseWriter, r *http.Request) {
	ttcs := CreateTaskStatistics(is.storage)

	is.getBasePage(
		components.TemplateTaskCompletionStatistics(ttcs),
		components.BaseComponentScripts(),
	).Component.Render(r.Context(), w)
}
