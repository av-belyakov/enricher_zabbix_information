package apiserver

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appversion"
	"github.com/av-belyakov/enricher_zabbix_information/internal/memorystatistics"
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

	unit := "hours"
	count := int(time.Since(is.timeStart).Hours())
	if count >= 48 {
		count = int(math.Floor(float64(count) / 24))
		unit = "days"
	}

	version := "v0.0.1"
	if v, err := appversion.GetVersion(); err != nil {
		version = "v" + v
	}

	io.WriteString(
		w,
		fmt.Sprintf("Hello, %s %s, application status:'%s'. %d %s have passed since the launch of the application.\n\n%s\n",
			appname.GetName(),
			version,
			status,
			count,
			unit,
			memorystatistics.PrintMemStats()))
}

// RouteTasks маршрут при обращении к '/tasks'
func (is *InformationServer) RouteTasks(w http.ResponseWriter, r *http.Request) {
	io.WriteString(
		w,
		"This is tasks' page!")
}
