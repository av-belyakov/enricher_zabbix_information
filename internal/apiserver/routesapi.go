package apiserver

import (
	"fmt"
	"io"
	"net/http"

	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
)

func (is *InformationServer) RouteApi(w http.ResponseWriter, r *http.Request) {
	defer supportingfunctions.CloseHTTPRequest(r)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		//http.Error(w, "this method is not supported", http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "")

		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		is.logger.Send("error", fmt.Sprintf("invalid http request accepted, learn more '%s'", err.Error()))
	}

	if r.URL.Query().Get("task_management") == "" {
		http.NotFound(w, r)

		return
	}

	if is.transmitterFromFrontend == nil {
		is.logger.Send("warning", "the transmitter for transmitting data from external sources is not initialized")

		return
	}

	// передаем данные пришедшие от внешнего источника
	is.transmitterFromFrontend.Send(b)

	fmt.Println("method 'InformationServer.RouteApi' received POST request, body:", string(b))

	fmt.Fprintf(w, "Success!")

}
