package apiserver

import (
	"net/http"

	"github.com/av-belyakov/enricher_zabbix_information/components"
)

func (is *InformationServer) RouteSSE(w http.ResponseWriter, r *http.Request) {
	//sse сервер
	go is.sseServer.HandleSSE(w, r)

	is.getBasePage(
		components.TemplateSSE(),
		components.BaseComponentScripts(),
	).Component.Render(r.Context(), w)
}
