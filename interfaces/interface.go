package interfaces

import (
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
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
}

type StorageInformation interface {
	GetStatusProcessRunning() bool
	GetDateExecution() (start, end time.Time)
	GetList() []storage.HostDetailedInformation
}

// ************** передача данных ***************
type BytesTransmitter interface {
	SendData([]byte)
	GetTypeTransmitter() string
	//ReceiveData() []byte
}
