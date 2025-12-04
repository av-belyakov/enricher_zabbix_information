package apiserver

import (
	"fmt"
	"net"
	"net/http"

	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
)

// RouteManuallyTaskStarting ручной запуск задачи
func (is *InformationServer) RouteManuallyTaskStarting(w http.ResponseWriter, r *http.Request) {
	var jsScripts string
	var msgTaskIsExecuting string
	if is.storage.GetStatusProcessRunning() {
		msgTaskIsExecuting = "<p>В настоящее время задача уже выполняется</p>"
	} else {
		jsScripts = fmt.Sprintf(`<script>
			let token = prompt("Введите токен авторизации для запуска задачи:");
			if (token) {
				const data = JSON.stringify({ token: token });
				const xhr = new XMLHttpRequest();
				xhr.open('POST', 'http://%s/api?task_management&task=start', true);
				xhr.setRequestHeader('Content-Type', 'application/json');
				xhr.send(data);
				xhr.onload = function() {
  					if (xhr.status != 200) { // анализируем HTTP-статус ответа, если статус не 200, то произошла ошибка
						alert('Ошибка ${xhr.status}: ${xhr.statusText}'); // Например, 404: Not Found
  					} else { // если всё прошло гладко, выводим результат
    					alert('Готово, получили ${xhr.response.length} байт'); // response -- это ответ сервера
  					}
				};
			}
		</script>`, net.JoinHostPort(is.host, fmt.Sprint(is.port)))
	}

	fmt.Fprintf(w, `
	  <!DOCTYPE html>
		<html lang="ru">
  		<head>
    		<meta charset="utf-8">
    		<meta name="viewport" content="width=device-width, initial-scale=1.0">
    		<title>%s</title>
			%s
		</head>
		<body>
			<h3>Запуск выполнения задачи вне расписания</h3>
			%s
    	</body>
		</html>
	`,
		appname.GetName(),
		jsScripts,
		msgTaskIsExecuting,
	)
}
