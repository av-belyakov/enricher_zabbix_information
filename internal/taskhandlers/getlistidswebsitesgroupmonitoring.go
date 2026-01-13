package taskhandlers

import (
	"slices"

	"github.com/av-belyakov/enricher_zabbix_information/internal/dictionarieshandler"
	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"
)

// GetListIdsWebsitesGroupMonitoring получение списка id групп хостов
// на основании словарей групп сайтов, по которым осуществляется мониторинг
func GetListIdsWebsitesGroupMonitoring(dictionaryPath string, hostGroupList []connectionjsonrpc.HostGroupInformation) ([]string, error) {
	var listGroupsId []string
	dicts, err := dictionarieshandler.Read(dictionaryPath)
	if err != nil {
		return listGroupsId, err
	}
	dictsSize := len(dicts.Dictionaries.WebSiteGroupMonitoring)

	for _, host := range hostGroupList {
		if dictsSize == 0 {
			listGroupsId = append(listGroupsId, host.GroupId)

			continue
		}

		if !slices.ContainsFunc(
			dicts.Dictionaries.WebSiteGroupMonitoring,
			func(v dictionarieshandler.WebSiteMonitoring) bool {
				return v.Name == host.Name
			},
		) {
			continue
		}

		listGroupsId = append(listGroupsId, host.GroupId)
	}

	return listGroupsId, nil
}
