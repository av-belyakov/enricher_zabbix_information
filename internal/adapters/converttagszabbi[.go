package adapters

import (
	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/internal/appstorage"
)

// ConvertTagsBetweenPackages конвертирует тэги между пакетами
func ConvertTagsBetweenPackages(oldTags connectionjsonrpc.Tags) []appstorage.Tag {
	tags := make([]appstorage.Tag, 0, len(oldTags.Tag))
	for _, v := range oldTags.Tag {
		tags = append(tags, appstorage.Tag{
			Tag:   v.Tag,
			Value: v.Value,
		})
	}

	return tags
}
