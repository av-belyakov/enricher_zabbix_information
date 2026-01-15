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
func (spl *ShortPrefixList) SearchIps(ips []netip.Addr) <-chan []ShortPrefixInfo {
	chanPrefixInfo := make(chan []ShortPrefixInfo)

	go func() {
		defer close(chanPrefixInfo)

		for _, ip := range ips {
			indexes := spl.SearchIp(ip)
			if len(indexes) == 0 {
				continue
			}

			var spi []ShortPrefixInfo
			for _, index := range indexes {
				spi = append(spi, spl.Prefixes[index])
			}

			chanPrefixInfo <- spi
		}
	}()

	return chanPrefixInfo
}

// SearchIp поиск IP-адреса в списке префиксов
func (spl *ShortPrefixList) SearchIp(ip netip.Addr) []int {
	var indexes []int
	left := 0
	right := len(spl.Prefixes) - 1

	spl.mutex.RLock()
	defer spl.mutex.RUnlock()

	for left <= right {
		if spl.Prefixes[left].Prefix.Contains(ip) {
			indexes = append(indexes, left)
		}

		if left != right && spl.Prefixes[right].Prefix.Contains(ip) {
			indexes = append(indexes, right)
		}

		left++
		right--
	}

	return indexes
}
