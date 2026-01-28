package apiserver

import (
	"net/http"
	"os"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/components"
	"github.com/av-belyakov/enricher_zabbix_information/constants"
)

// RouteIndex маршрут при обращении к '/'
func (is *InformationServer) RouteIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	appStatus := "production"
	if os.Getenv("GO_"+constants.App_Environment_Name+"_MAIN") == "development" ||
		os.Getenv("GO_"+constants.App_Environment_Name+"_MAIN") == "test" {
		appStatus = os.Getenv("GO_" + constants.App_Environment_Name + "_MAIN")
	}

	appTimeLive := time.Since(is.timeStart).String()

	is.getBasePage(
		components.TemplateMainElement(appStatus, appTimeLive, is.storage),
		components.BaseComponentScripts(),
	).Component.Render(r.Context(), w)
}
