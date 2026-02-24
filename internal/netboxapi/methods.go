package netboxapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
)

// Get реализация HTTP GET запроса
func (api *Client) Get(ctx context.Context, query string) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(api.settings.timeout)*time.Second)
	defer cancel()

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

// Len длина списка
func (spl *ShortPrefixList) Len() int {
	return len(*spl)
}

// SearchIps поиск IP-адресов в списке префиксов
func (spl *ShortPrefixList) SearchIps(ips []netip.Addr) <-chan []ShortPrefixInfo {
	chanPrefixInfo := make(chan []ShortPrefixInfo)

	go func() {
		defer close(chanPrefixInfo)

		for _, ip := range ips {
			list := spl.SearchIp(ip)
			if len(list) == 0 {
				continue
			}

			chanPrefixInfo <- list
		}
	}()

	return chanPrefixInfo
}

// SearchIp поиск IP-адреса в списке префиксов
func (spl *ShortPrefixList) SearchIp(ip netip.Addr) []ShortPrefixInfo {
	var list []ShortPrefixInfo
	left := 0
	right := len(*spl) - 1

	for left <= right {
		if (*spl)[left].Prefix.Contains(ip) {
			list = append(list, (*spl)[left])
		}

		if left != right && (*spl)[right].Prefix.Contains(ip) {
			list = append(list, (*spl)[right])
		}

		left++
		right--
	}

	return list
}
