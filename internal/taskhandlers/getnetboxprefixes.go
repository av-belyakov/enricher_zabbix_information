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

func NetboxPrefixes(ctx context.Context, client *netboxapi.Client, logger interfaces.Logger) (<-chan []netboxapi.ShortPrefixInfo, int, error) {
	chOut := make(chan []netboxapi.ShortPrefixInfo)

	// выясняем сколько всего префиксов
	res, statusCode, err := client.Get(ctx, "/api/ipam/prefixes/?limit=1")
	if err != nil {
		logger.Send("error", wrappers.WrapperError(err).Error())

		return chOut, 0, err
	}

	if statusCode != http.StatusOK {
		logger.Send("error", wrappers.WrapperError(fmt.Errorf("status code %d was received", statusCode)).Error())

		return chOut, 0, err
	}

	var nbPrefixes netboxapi.ListPrefixes
	err = json.Unmarshal(res, &nbPrefixes)
	if err != nil {
		logger.Send("error", wrappers.WrapperError(err).Error())

		return chOut, 0, err
	}

	chunkSize := 100
	if nbPrefixes.Count > 1_000 {
		chunkSize = 500
	}

	chunkCount := math.Ceil(float64(nbPrefixes.Count) / float64(chunkSize))

	go func() {
		defer close(chOut)

		// получаем полный список префиксов по частям
		for i := 0; i < int(chunkCount); i++ {
			prefixList := make([]netboxapi.ShortPrefixInfo, 0)

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

				prefixList = append(prefixList, data.Information)
			}

			chOut <- prefixList
		}
	}()

	return chOut, nbPrefixes.Count, nil
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
