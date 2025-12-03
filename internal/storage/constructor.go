package storage

import "github.com/av-belyakov/enricher_zabbix_information/datamodels"

// NewShortTermStorage конструктор кратковременного хранилища
func NewShortTermStorage() *ShortTermStorage {
	return &ShortTermStorage{
		data: []datamodels.HostDetailedInformation{},
	}
}
