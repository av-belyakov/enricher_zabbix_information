package storage

import (
	"sync"
	"sync/atomic"

	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
)

// ShortTermStorage хранилище в памяти
type ShortTermStorage struct {
	data        []datamodels.HostDetailedInformation
	mutex       sync.RWMutex
	isExecution atomic.Bool
}
