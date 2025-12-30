package dnsresolver

import (
	"net"
	"net/netip"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// Settings настройки
type Settings struct {
	resolver *net.Resolver
	storage  interfaces.StorageDNSResolver
	timeout  time.Duration
}

type Options func(*Settings) error

type InfoFromDNSResolver struct {
	Ips          []netip.Addr
	Error        error
	OriginalHost string
	DomainName   string
	HostId       int
}
