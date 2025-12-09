package interfaces

import (
	"net/netip"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
)

// **************** счётчик *****************
type Counter interface {
	SendMessage(string, int)
}

// ************** логирование ***************
type Logger interface {
	GetChan() <-chan Messager
	Send(msgType, msgData string)
}

type Messager interface {
	GetType() string
	SetType(v string)
	GetMessage() string
	SetMessage(v string)
}

type WriterLoggingData interface {
	Write(typeLogFile, str string) bool
}

// ************** хранилище ***************
type StorageDNSResolver interface {
	GetHosts() map[int]string
	SetIps(hostId int, ip netip.Addr, ips ...netip.Addr) error
	SetError(hostId int, err error) error
	SetDomainName(hostId int, domainName string) error
}

type StorageInformation interface {
	GetStatusProcessRunning() bool
	GetDateExecution() (start, end time.Time)
	GetList() []datamodels.HostDetailedInformation
}

// ************** передача данных ***************
type BytesTransmitter interface {
	Send([]byte)
	Receive() []byte
}
