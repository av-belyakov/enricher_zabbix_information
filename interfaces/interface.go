package interfaces

import "net/netip"

//**************** счётчик *****************

type Counter interface {
	SendMessage(string, int)
}

//************** логирование ***************

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

// ************** хранилище преобразователя доменных имён ***************
type StorageDNSResolver interface {
	GetHosts() map[int]string
	SetIps(hostId int, ip netip.Addr, ips ...netip.Addr) error
	SetError(hostId int, err error) error
	SetDomainName(hostId int, domainName string) error
}
