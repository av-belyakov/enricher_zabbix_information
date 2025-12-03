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
	if is.version != "" {
		version = "v" + is.version
	}

	hellowMsg := fmt.Sprintf(
		"Hello, %s %s, application status:'%s'. %d %s have passed since the launch of the application.\n\n\n",
		appname.GetName(),
		version,
		status,
		count,
		unit,
	)

	fmt.Fprintf(w, `
	  <!DOCTYPE html>
		<html lang="ru">
  		<head>
    		<meta charset="utf-8">
    		<meta name="viewport" content="width=device-width, initial-scale=1.0">
    		<title>%s</title>
		</head>
		<body>
		 	<p>%s</p>
      		<nav>
        		<ul>
				<li><a href="manually_task_starting">ручной запуск задачи</a></li>
          			<li><a href="task_information">информация о выполненной задаче</a></li>
          			<li><a href="memory_statistics">общая статистика расходования памяти</a></li>
        		</ul>
      		</nav>
    	</body>
		</html>
	`,
		appname.GetName(),
		hellowMsg,
	)
}

// RouteMemoryStatistics статистика использования памяти
func (is *InformationServer) RouteMemoryStatistics(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintf("Memory statistics:\n%s\n", memorystatistics.PrintMemStats()))
}

// RouteTaskInformation информация о выполненной задаче
func (is *InformationServer) RouteTaskInformation(w http.ResponseWriter, r *http.Request) {
	io.WriteString(
		w,
		"This is page with taks information!")
}

// RouteManuallyTaskStarting ручной запуск задачи
func (is *InformationServer) RouteManuallyTaskStarting(w http.ResponseWriter, r *http.Request) {
	io.WriteString(
		w,
		"This is page with manually starting a task!")
}
