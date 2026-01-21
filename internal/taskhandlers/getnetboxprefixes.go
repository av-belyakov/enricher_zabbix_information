package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/netip"

	"github.com/bytedance/sonic"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/netboxapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

// GetNetboxPrefixes получить все nNetbox префиксы
func GetNetboxPrefixes(ctx context.Context, client *netboxapi.Client, logger interfaces.Logger) *netboxapi.ShortPrefixList {
	shortPrefixList := &netboxapi.ShortPrefixList{}

	// выясняем сколько всего префиксов
	res, statusCode, err := client.Get(ctx, "/api/ipam/prefixes/?limit=1")
	if err != nil {
		logger.Send("error", wrappers.WrapperError(err).Error())

		return shortPrefixList
	}

	if statusCode != http.StatusOK {
		logger.Send("error", wrappers.WrapperError(fmt.Errorf("status code %d was received", statusCode)).Error())

		return shortPrefixList
	}

	var nbPrefixes netboxapi.ListPrefixes
	err = json.Unmarshal(res, &nbPrefixes)
	if err != nil {
		logger.Send("error", wrappers.WrapperError(err).Error())

		return shortPrefixList
	}

	shortPrefixList.Count = nbPrefixes.Count

	chunkSize := 100
	if nbPrefixes.Count > 1_000 {
		chunkSize = 500
	}

	chunkCount := math.Ceil(float64(nbPrefixes.Count) / float64(chunkSize))

	// получаем полный список префиксов по частям
	for i := 0; i < int(chunkCount); i++ {
		offset := i * chunkSize
		res, statusCode, err := client.Get(ctx, fmt.Sprintf("/api/ipam/prefixes/?limit=%d&offset=%d", chunkSize, offset))
		if err != nil {
			logger.Send("error", wrappers.WrapperError(err).Error())

			continue
		}
		if statusCode != http.StatusOK {
			logger.Send("error", wrappers.WrapperError(fmt.Errorf("status code %d was received", statusCode)).Error())
		}

		nbPrefixes := netboxapi.ListPrefixes{}
		err = sonic.Unmarshal(res, &nbPrefixes)
		if err != nil {
			logger.Send("error", wrappers.WrapperError(err).Error())

			continue
		}

		for data := range getShortPrefixesInformation(nbPrefixes) {
			if data.Error != nil {
				logger.Send("error", wrappers.WrapperError(data.Error).Error())

				continue
			}

			shortPrefixList.Prefixes = append(shortPrefixList.Prefixes, data.Information)
		}
	}

	return shortPrefixList
}

func getShortPrefixesInformation(prefixes netboxapi.ListPrefixes) <-chan struct {
	Information netboxapi.ShortPrefixInfo
	Error       error
} {
	chanInfo := make(chan struct {
		Information netboxapi.ShortPrefixInfo
		Error       error
	})

	go func() {
		defer close(chanInfo)

		for _, prefix := range prefixes.Results {
			netPrefix, err := netip.ParsePrefix(prefix.Prefix)
			if err != nil {
				chanInfo <- struct {
					Information netboxapi.ShortPrefixInfo
					Error       error
				}{
					Error: err,
				}

				continue
			}

			var sensorId string
			if len(prefix.CustomFields.Sensors) > 0 {
				sensorId = prefix.CustomFields.Sensors[0].Name
			}

			chanInfo <- struct {
				Information netboxapi.ShortPrefixInfo
				Error       error
			}{
				Information: netboxapi.ShortPrefixInfo{
					Status:   prefix.Status.Value,
					Prefix:   netPrefix,
					Id:       prefix.Id,
					SensorId: sensorId,
				},
			}
		}
	}()

	return chanInfo
}
