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
}

type Options func(*Settings) error
