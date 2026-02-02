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

func NetboxPrefixes(ctx context.Context, client *netboxapi.Client, logger interfaces.Logger) (<-chan netboxapi.ShortPrefixList /*[]netboxapi.ShortPrefixInfo*/, int, error) {
	chOut := make(chan netboxapi.ShortPrefixList)

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

	var maxCountPrefixes netboxapi.ListPrefixes
	err = json.Unmarshal(res, &maxCountPrefixes)
	if err != nil {
		logger.Send("error", wrappers.WrapperError(err).Error())

		return chOut, 0, err
	}

	chunkSize := 100
	if maxCountPrefixes.Count > 1_000 {
		chunkSize = 500
	}
	if maxCountPrefixes.Count > 1_500 {
		chunkSize = 750
	}

	chunkCount := math.Ceil(float64(maxCountPrefixes.Count) / float64(chunkSize))

	go func() {
		defer close(chOut)

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

			netboxPrefixes := netboxapi.ListPrefixes{}
			err = sonic.Unmarshal(res, &netboxPrefixes)
			if err != nil {
				logger.Send("error", wrappers.WrapperError(err).Error())

				continue
			}

			shortPrefixList := make(netboxapi.ShortPrefixList, 0, netboxPrefixes.Count)
			for data := range getShortPrefixesInformation(netboxPrefixes) {
				if data.Error != nil {
					logger.Send("error", wrappers.WrapperError(data.Error).Error())

					continue
				}

				shortPrefixList = append(shortPrefixList, data.Information)
			}

			chOut <- shortPrefixList
		}
	}()

	return chOut, maxCountPrefixes.Count, nil
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
