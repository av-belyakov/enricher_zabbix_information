package taskhandlers

import (
	"cmp"
	"sync"

	"github.com/av-belyakov/enricher_zabbix_information/internal/appstorage"
	"github.com/av-belyakov/enricher_zabbix_information/internal/netboxapi"
)

// SearchIpToNetboxPrefixes поиск списка ip адресов в префиксах netbox
func SearchIpToNetboxPrefixes(hosts []appstorage.HostDetailedInformation, chanInput <-chan netboxapi.ShortPrefixList) chan SearchResponse {
	chanOutput := make(chan SearchResponse)

	go func() {
		var wg sync.WaitGroup

		for v := range chanInput {
			// это нужно что бы передать количество принятых из Netbox перфиксов
			chanOutput <- SearchResponse{
				SizeProcessedList: v.Len(),
			}

			wg.Go(func() {
				for _, host := range hosts {
					for foundInfo := range v.SearchIps(host.Ips) {
						for _, item := range foundInfo {
							chanOutput <- SearchResponse{
								SearchDetailedInformation: DetailedInformation{
									HostId:   host.HostId,
									NetboxId: item.Id,
									SensorId: item.SensorId,
									IsActive: cmp.Or(item.Status == "active", true, false),
								},
							}
						}
					}
				}
			})
		}

		wg.Wait()
		close(chanOutput)
	}()

	return chanOutput
}
