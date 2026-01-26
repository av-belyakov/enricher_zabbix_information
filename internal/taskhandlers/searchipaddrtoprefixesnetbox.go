package taskhandlers

import (
	"runtime"
	"sync"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appstorage"
	"github.com/av-belyakov/enricher_zabbix_information/internal/netboxapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

// SearchIpaddrToPrefixesNetbox поиск списка ip адресов в префиксах netbox
func SearchIpaddrToPrefixesNetbox(
	maxCountGoroutines int,
	storageTemp *appstorage.SharedAppStorage,
	shortPrefixList *netboxapi.ShortPrefixList,
	logging interfaces.Logger,
) {
	goroutines := 3
	chQueue := make(chan appstorage.HostDetailedInformation, maxCountGoroutines)

	if maxCountGoroutines >= 1 && maxCountGoroutines <= runtime.NumCPU() {
		maxCountGoroutines = goroutines
	}

	storageTempSetInfo := func(id int, info []netboxapi.ShortPrefixInfo, st *appstorage.SharedAppStorage) error {
		if err := st.SetIsProcessed(id); err != nil {
			return err
		}

		for _, msg := range info {
			if msg.Status != "active" {
				continue
			}

			st.SetIsActive(id)
			st.SetSensorId(id, msg.SensorId)
			st.SetNetboxHostId(id, msg.Id)
		}

		return nil
	}

	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for hostInfo := range chQueue {
				for msgList := range shortPrefixList.SearchIps(hostInfo.Ips) {
					if err := storageTempSetInfo(hostInfo.HostId, msgList, storageTemp); err != nil {
						logging.Send("error", wrappers.WrapperError(err).Error())
					}
				}
			}
		})
	}

	for _, v := range storageTemp.GetList() {
		chQueue <- v
	}

	close(chQueue)
	wg.Wait()
}
