package dnsresolver

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/av-belyakov/enricher_zabbix_information/test/helpersfile"
)

func TestDnsResolver(t *testing.T) {
	b, err := os.ReadFile("../filesfortest/exampledata.json")
	if err != nil {
		log.Fatalln(err)
	}

	examleData := helpersfile.TypeExampleData{}
	if err := json.Unmarshal(b, &examleData); err != nil {
		log.Fatalln(err)
	}
	if len(examleData.Hosts) == 0 {
		log.Fatalln(errors.New("the structure 'TypeExampleData' should not be empty"))
	}

	sts := storage.NewShortTermStorage()

	//наполняем хранилище
	for _, v := range examleData.Hosts {
		hostId, err := strconv.Atoi(v.HostId)
		assert.NoError(t, err)

		sts.Add(storage.HostDetailedInformation{
			HostId:       hostId,
			OriginalHost: v.Host,
		})
	}

	listElement := sts.GetList()
	if len(listElement) == 0 {
		log.Fatalln(errors.New("the storage should not be empty"))
	}

	t.Run("Тест 1. Выполняем верификацию доменных имён. ", func(t *testing.T) {
		for _, v := range listElement {
			urlHost, err := url.Parse(v.OriginalHost)
			if err != nil {
				t.Errorf(
					"Host %s is not valid URL. Error: %s",
					v.OriginalHost,
					err.Error(),
				)
			}
		}
	})

}
