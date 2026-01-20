package dnsresolver

import (
	"context"
	"errors"
	"net/netip"
	"net/url"
	"strings"

	"github.com/av-belyakov/enricher_zabbix_information/internal/customerrors"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

// Run запуск преобразования списка доменных имён в ip адреса
func (s *Settings) Run(ctx context.Context, hosts map[int]string) (<-chan InfoFromDNSResolver, error) {
	chSendData := make(chan InfoFromDNSResolver)

	if len(hosts) == 0 {
		return nil, wrappers.WrapperError(errors.New("the list of domain names intended for searching ip addresses should not be empty"))
	}

	go func() {
		defer close(chSendData)

		for hostId, originalHost := range hosts {
			idns := InfoFromDNSResolver{
				HostId:       hostId,
				OriginalHost: originalHost,
			}

			if !strings.Contains(originalHost, "http://") {
				originalHost = "http://" + originalHost
			}

			urlHost, err := url.Parse(originalHost)
			if err != nil {
				idns.Error = customerrors.NewErrorNoValidUrl(originalHost, err)
				chSendData <- idns

				continue
			}

			idns.DomainName = urlHost.Host

			// DNS resolve
			ips, err := s.resolver.LookupHost(ctx, urlHost.Host)
			if err != nil {
				idns.Error = customerrors.NewErrorUrlNotFound(originalHost, err)
				chSendData <- idns

				continue
			}

			if len(ips) == 0 {
				idns.Error = customerrors.NewErrorUrlNotFound(originalHost, err)
				chSendData <- idns

				continue
			}

			hostIps := make([]netip.Addr, 0, len(ips))
			for _, ip := range ips {
				ipaddr, err := netip.ParseAddr(ip)
				if err != nil {
					idns.Error = customerrors.NewErrorIpInvalid(ip, err)
					chSendData <- idns

					continue
				}

				hostIps = append(hostIps, ipaddr)
			}

			idns.Ips = hostIps
			chSendData <- idns
		}
	}()

	return chSendData, nil
}
