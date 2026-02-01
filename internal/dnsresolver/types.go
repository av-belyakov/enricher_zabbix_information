package dnsresolver

import (
	"net"
	"net/netip"
	"time"
)

// Settings настройки
type Settings struct {
	resolver *net.Resolver
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
