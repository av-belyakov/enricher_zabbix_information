package apiserver

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

	version := "v0.0.1"
	if is.version != "" {
		version = "v" + is.version
	}

	appName := fmt.Sprintf(
		"%s %s\n",
		strings.ToTitle(appname.GetName()),
		version,
	)

	hellowMsg := fmt.Sprintf(
		"Статус приложения: '%s'. Прошло с момента запуска приложения: %s.",
		status,
		time.Since(is.timeStart).String(),
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
		appName,
		hellowMsg,
	)
}

// RouteMemoryStatistics статистика использования памяти
func (is *InformationServer) RouteMemoryStatistics(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintf("Memory statistics:\n%s\n", memorystatistics.PrintMemStats()))
}

// RouteTaskInformation информация о выполненной задаче
func (is *InformationServer) RouteTaskInformation(w http.ResponseWriter, r *http.Request) {
	taskStatus := "завершена"
	if is.storage.GetStatusProcessRunning() {
		taskStatus = "выполняется"
	}

	start, end := is.storage.GetDateExecution()
	dataStart := start.Format("15:04:05 2006-01-02")
	dataEnd := "-"
	diffTime := "-"
	if end.String() != "00:00:00 0001-01-01" {
		dataEnd = end.Format("15:04:05 2006-01-02")
		diffTime = end.Sub(start).String()
	}

	var countHostsError int
	hosts := is.storage.GetList()
	listHostsError := strings.Builder{}
	for _, v := range hosts {
		if v.Error == nil {
			continue
		}

		countHostsError++
		listHostsError.WriteString(
			fmt.Sprintf(
				"<li><b>%s</b>, error: %s</li>",
				v.OriginalHost,
				v.Error.Error(),
			))
	}

	fmt.Fprintf(w, `
	  <!DOCTYPE html>
		<html lang="ru">
  		<head>
    		<meta charset="utf-8">
    		<meta name="viewport" content="width=device-width, initial-scale=1.0">
    		<title>%s</title>
		</head>
		<body>
			<h3>Статистика выполнения задачи:</h1>		
			<div>Статус задачи: <u>%s</u></div>
			<div>Время начала выполнения: %s</div>
			<div>Время завершения выполнения: %s</div>
			<div>Время на выполнение задачи: %s</div>
			<div>Количество обработанных доменных имён: %d</div>
			<div>Количество доменных имён обработанных с ошибкой: %d</div>
			<div>Список доменных имён при обработки которых возникли ошибки:</div>
			<div>
				<ol>%s</ol>
			</div>
    	</body>
		</html>
	`,
		appname.GetName(),
		taskStatus,
		dataStart,
		dataEnd,
		diffTime,
		len(hosts),
		countHostsError,
		listHostsError.String(),
	)

	/*
			- задача выполняется или выполнилась
			- время начала задачи
			- время окончания задачи
			- сколько потребовалось времение на выполнения задачи
			- кол-во доменных имен обработано
		 	- кол-во доменных имен обработанных с ошибкой
				(список этих доменных имён с кратким описанием ошибки)

	*/
}

// RouteManuallyTaskStarting ручной запуск задачи
func (is *InformationServer) RouteManuallyTaskStarting(w http.ResponseWriter, r *http.Request) {
	io.WriteString(
		w,
		"This is page with manually starting a task!")
}
