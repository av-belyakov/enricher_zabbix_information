package apiserver

import (
	"fmt"
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

	status := "production"
	if os.Getenv("GO_"+constants.App_Environment_Name+"_MAIN") == "development" ||
		os.Getenv("GO_"+constants.App_Environment_Name+"_MAIN") == "test" {
		status = os.Getenv("GO_" + constants.App_Environment_Name + "_MAIN")
	}

	hellowMsg := fmt.Sprintf(
		"Статус приложения: '%s'. Прошло времени с момента запуска приложения: %s.",
		status,
		time.Since(is.timeStart).String(),
	)

	is.getBasePage(
		components.MainElement(hellowMsg),
		components.BaseComponentScripts(),
	).Component.Render(r.Context(), w)
}
