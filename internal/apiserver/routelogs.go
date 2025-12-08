package apiserver

import (
	"net/http"

	"github.com/av-belyakov/enricher_zabbix_information/components"
)

func (is *InformationServer) RouteLogs(w http.ResponseWriter, r *http.Request) {
	is.getBasePage(
		components.TemplateLogs(),
		components.BaseComponentScripts(),
	).Component.Render(r.Context(), w)
}
