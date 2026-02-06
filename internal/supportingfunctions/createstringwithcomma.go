package supportingfunctions

import (
	"net/netip"
	"strings"
)

// CreateStringWithComma создание строки с разделителем запятая
func CreateStringWithComma(list []string) string {
	if len(list) == 0 {
		return ""
	}

	if len(list) == 1 {
		return list[0]
	}

	return strings.Join(list, ", ")
}

// CreateStringWithCommaFromIps создание строки с разделителем запятая из IP-адресов
func CreateStringWithCommaFromIps(ips []netip.Addr) string {
	if len(ips) == 0 {
		return ""
	}

	if len(ips) == 1 {
		return ips[0].String()
	}

	list := make([]string, 0, len(ips))
	for _, ip := range ips {
		list = append(list, ip.String())
	}

	return strings.Join(list, ", ")
}
