package supportingfunctions

import "net/http"

// CloseHTTPRequest закрытие http соединения с предварительной проверкой
func CloseHTTPRequest(req *http.Request) {
	if req == nil || req.Body == nil {
		return
	}

	req.Body.Close()
}

// CloseHTTPResponse закрытие http соединения с предварительной проверкой
func CloseHTTPResponse(res *http.Response) {
	if res == nil || res.Body == nil {
		return
	}

	res.Body.Close()
}
