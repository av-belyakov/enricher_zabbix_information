package dnsresolver

import (
	"context"
	"errors"
	"net/netip"
	"net/url"
	"strings"

	"golang.org/x/net/idna"

	"github.com/av-belyakov/enricher_zabbix_information/internal/customerrors"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

// Run запуск преобразования списка доменных имён в ip адреса
func (s *Settings) Run(ctx context.Context, hosts []ShortInformationAboutHost) (<-chan InfoFromDNSResolver, error) {
	if len(hosts) == 0 {
		return nil, wrappers.WrapperError(errors.New("the list of domain names intended for searching ip addresses should not be empty"))
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	chSendData := make(chan InfoFromDNSResolver)

	go func() {
		defer func() {
			cancel()
			close(chSendData)
		}()

		for _, v := range hosts {
			idns := InfoFromDNSResolver{
				HostId:       v.GetHostId(),
				OriginalHost: v.GetOriginalHost(),
			}

			if !strings.HasPrefix(v.GetOriginalHost(), "http") {
				idns.Error = customerrors.NewErrorNoValidUrl(idns.OriginalHost, "the http prefix is missing")
				idns.OriginalHost = "http://" + v.GetOriginalHost()
			}

			urlHost, err := url.ParseRequestURI(idns.OriginalHost)
			if err != nil {
				idns.Error = customerrors.NewErrorNoValidUrl(idns.OriginalHost, err.Error())
				chSendData <- idns

				continue
			}

			asciiDomain, err := idna.ToASCII(urlHost.Hostname())
			if err != nil {
				idns.Error = customerrors.NewErrorNoValidUrl(idns.OriginalHost, err.Error())
				chSendData <- idns

				continue
			}

			idns.DomainName = asciiDomain

			// DNS resolve
			ips, err := s.resolver.LookupHost(ctx, idns.DomainName)
			if err != nil {
				idns.Error = customerrors.NewErrorUrlNotFound(idns.OriginalHost, err.Error())
				chSendData <- idns

				continue
			}

			if len(ips) == 0 {
				idns.Error = customerrors.NewErrorUrlNotFound(idns.OriginalHost, "")
				chSendData <- idns

				continue
			}

			hostIps := make([]netip.Addr, 0, len(ips))
			for _, ip := range ips {
				ipaddr, err := netip.ParseAddr(ip)
				if err != nil {
					idns.Error = customerrors.NewErrorIpInvalid(ip, err.Error())
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
