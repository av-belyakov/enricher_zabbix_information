package storage

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
)

// ShortTermStorage хранилище в памяти
type ShortTermStorage struct {
	data               []datamodels.HostDetailedInformation
	mutex              sync.RWMutex
	startDateExecution time.Time
	endDateExecution   time.Time
	isExecution        atomic.Bool
}
