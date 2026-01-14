package netboxapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/netip"

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

// SearchIps поиск IP-адресов в списке префиксов
func (spl *ShortPrefixList) SearchIps(ips []netip.Addr) <-chan ShortPrefixInfo {
	chanPrefixInfo := make(chan ShortPrefixInfo)

	go func() {
		defer close(chanPrefixInfo)

		for _, ip := range ips {
			if index, ok := spl.SearchIp(ip); ok {
				chanPrefixInfo <- spl.Prefixes[index]
			}
		}
	}()

	return chanPrefixInfo
}

// SearchIp поиск IP-адреса в списке префиксов
func (spl *ShortPrefixList) SearchIp(ip netip.Addr) (int, bool) {
	spl.mutex.RLock()
	defer spl.mutex.RUnlock()

	/*
		for index, prefix := range spl.Prefixes {
			if prefix.Prefix.Contains(ip) {
				return index, true
			}
		}
	*/

	left := 0
	right := len(spl.Prefixes) - 1

	for left <= right {
		if spl.Prefixes[left].Prefix.Contains(ip) {
			return left, true
		}

		if left != right && spl.Prefixes[right].Prefix.Contains(ip) {
			return right, true
		}

		left++
		right--
	}

	return -1, false
}
