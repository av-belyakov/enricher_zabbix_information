package dnsresolver

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"strings"

	"github.com/av-belyakov/enricher_zabbix_information/internal/customerrors"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

// Run запуск преобразования списка доменных имён в ip адреса
func (s *Settings) Run(ctx context.Context, chFinish chan<- struct{}) {
	fmt.Printf("___ method 'Settings.Run' s.chSignal='%v' (type:'%T')\n", s.chSignal, s.chSignal)

	hostList := s.storage.GetHosts()

	if len(hostList) == 0 {
		s.logger.Send("error", wrappers.WrapperError(errors.New("the list of domain names intended for searching ip addresses should not be empty")).Error())
		chFinish <- struct{}{}

		return
	}

	for hostId, originalHost := range hostList {
		if err := s.storage.SetIsProcessed(hostId); err != nil {
			s.logger.Send("error", wrappers.WrapperError(err).Error())
		}

		if !strings.Contains(originalHost, "http://") {
			originalHost = "http://" + originalHost
		}

		urlHost, err := url.Parse(originalHost)
		if err != nil {
			if err := s.storage.SetError(hostId, customerrors.NewErrorNoValidUrl(originalHost, err)); err != nil {
				s.logger.Send("error", wrappers.WrapperError(err).Error())
			}

			// сообщение в канал для информирования внешних систем о произошедших в хранилище изменениях
			if s.chSignal != nil {
				fmt.Println("111")

				s.chSignal <- struct{}{}
			}

			continue
		}

		// DNS resolve
		ips, err := s.resolver.LookupHost(ctx, urlHost.Host)
		if err != nil {
			if err := s.storage.SetError(hostId, customerrors.NewErrorUrlNotFound(originalHost, err)); err != nil {
				s.logger.Send("error", wrappers.WrapperError(err).Error())
			}

			if s.chSignal != nil {
				fmt.Println("222")

				s.chSignal <- struct{}{}
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

			if s.chSignal != nil {
				fmt.Println("333")

				s.chSignal <- struct{}{}
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

				if s.chSignal != nil {
					fmt.Println("444")

					s.chSignal <- struct{}{}
				}

				continue
			}

			hostIps = append(hostIps, ipaddr)
		}

		if err := s.storage.SetIps(hostId, hostIps[0], hostIps...); err != nil {
			s.logger.Send("error", wrappers.WrapperError(err).Error())
		}

		if s.chSignal != nil {
			fmt.Println("555")

			s.chSignal <- struct{}{}
		}
	}

	chFinish <- struct{}{}
}

// AddChanSignal канал для информирования внешних систем о произошедших
// изменениях внутри модуля
func (s *Settings) AddChanSignal(ch chan<- struct{}) {
	s.chSignal = ch
}
