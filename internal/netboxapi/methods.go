package netboxapi

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
)

// Get реализация HTTP GET запроса
func (api *Client) Get(ctx context.Context, query string) ([]byte, int, error) {
	url := fmt.Sprintf("http://%s:%d%s", api.settings.host, api.settings.port, query)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Add("Authorization", "Token "+api.settings.token)
	req.Header.Set("Content-Type", "application/json")

	res, err := api.client.Do(req)
	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, fmt.Errorf("status code: %d (%s)", res.StatusCode, res.Status)
	}
	defer supportingfunctions.CloseHTTPResponse(res)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 500, err
	}

	return resBody, res.StatusCode, nil
}
