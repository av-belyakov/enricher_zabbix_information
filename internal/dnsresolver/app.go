package dnsresolver

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"net/url"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/customerrors"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

// New конструктор
func New(storage interfaces.StorageDNSResolver, logger interfaces.Logger, opts ...Options) (*Settings, error) {
	settings := &Settings{
		logger:  logger,
		storage: storage,
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, address)
			},
		},
		timeout: time.Second * 15,
	}

	for _, opt := range opts {
		if err := opt(settings); err != nil {
			return settings, err
		}
	}

	return settings, nil
}

// Run запуск преобразования списка доменных имён в ip адреса
func (s *Settings) Run(ctx context.Context, chFinish chan<- struct{}) {
	hostList := s.storage.GetHosts()

	if len(hostList) == 0 {
		s.logger.Send("error", wrappers.WrapperError(errors.New("the list of domain names intended for searching ip addresses should not be empty")).Error())
		chFinish <- struct{}{}

		return
	}

	for hostId, originalHost := range hostList {
		urlHost, err := url.Parse("http://" + originalHost)
		if err != nil {
			if err := s.storage.SetError(hostId, customerrors.NewErrorNoValidUrl(originalHost, err)); err != nil {
				s.logger.Send("error", wrappers.WrapperError(err).Error())
			}

			continue
		}

		ips, err := s.resolver.LookupHost(ctx, urlHost.Host)
		if err != nil {
			if err := s.storage.SetError(hostId, customerrors.NewErrorUrlNotFound(originalHost, err)); err != nil {
				s.logger.Send("error", wrappers.WrapperError(err).Error())
			}

			continue
		}

		if err := s.storage.SetDomainName(hostId, urlHost.Host); err != nil {
			s.logger.Send("error", wrappers.WrapperError(err).Error())
		}

		if len(ips) == 0 {
			if err := s.storage.SetError(hostId, customerrors.NewErrorUrlNotFound(urlHost.Host, err)); err != nil {
				s.logger.Send("error", wrappers.WrapperError(err).Error())
			}

			continue
		}

		hostIps := make([]netip.Addr, 0, len(ips))
		for _, ip := range ips {
			ipaddr, err := netip.ParseAddr(ip)
			if err != nil {
				if err := s.storage.SetError(hostId, customerrors.NewErrorIpInvalid(ip, err)); err != nil {
					s.logger.Send("error", wrappers.WrapperError(err).Error())
				}

				continue
			}

			hostIps = append(hostIps, ipaddr)
		}

		if err := s.storage.SetIps(hostId, hostIps[0], hostIps...); err != nil {
			s.logger.Send("error", wrappers.WrapperError(err).Error())
		}
	}

	chFinish <- struct{}{}
}
