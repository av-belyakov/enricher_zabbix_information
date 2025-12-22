package dnsresolver

import (
	"net"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// Settings настройки
type Settings struct {
	resolver *net.Resolver
	storage  interfaces.StorageDNSResolver
	logger   interfaces.Logger
	timeout  time.Duration
	chSignal chan<- struct{} // канал информирующий о произошедших изменениях внутри модуля
}

type Options func(*Settings) error
