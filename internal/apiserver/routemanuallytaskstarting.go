package apiserver

import (
	"net/http"

	"github.com/av-belyakov/enricher_zabbix_information/components"
)

// RouteManuallyTaskStarting ручной запуск задачи
func (is *InformationServer) RouteManuallyTaskStarting(w http.ResponseWriter, r *http.Request) {
	is.getBasePage(
		components.TemplateManuallyTaskStarting(is.storage.GetStatusProcessRunning()),
		components.BaseComponentScripts(),
	).Component.Render(r.Context(), w)
}
