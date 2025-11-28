package dnsresolver

import (
	"encoding/json"
	"errors"
	"fmt"
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

	t.Run("Тест 0. Просто проверка url", func(t *testing.T) {
		urlHost, err := url.Parse("http11://" + "argus.vetrf.ru")
		assert.NoError(t, err)

		fmt.Printf("urlHost 1 '%+v'\n", urlHost)
		fmt.Printf("urlHost 1 '%+v'\n", urlHost.Host)

		urlHost, err = url.Parse("http://" + "www.vetrf.ruvetrfcerberus.html")
		assert.NoError(t, err)

		fmt.Printf("urlHost 2 '%+v'\n", urlHost)
		fmt.Printf("urlHost 2 '%+v'\n", urlHost.Host)

		urlHost, err = url.Parse("http://" + "www.xn----7sbhhgjt4afav0m.xn--p1ai")
		assert.NoError(t, err)

		fmt.Printf("urlHost 3 '%+v'\n", urlHost)
		fmt.Printf("urlHost 3 '%+v'\n", urlHost.Host)
	})

	/*t.Run("Тест 1. Выполняем верификацию доменных имён.", func(t *testing.T) {
		for _, v := range listElement {
			urlHost, err := url.Parse(v.OriginalHost)
			if err != nil {
				fmt.Println("ERROR:", err)

				sts.SetError(v.HostId, customerrors.NewErrorNoValidUrl(v.OriginalHost))
			}

			fmt.Printf("v.OriginalHost:'%s', urlHost.Host:'%s'\n", v.OriginalHost, urlHost.Host)

			assert.NoError(t, sts.SetDomainName(v.HostId, urlHost.Host))
		}

		errList := sts.GetListErrors()
		fmt.Println("Element with errors:", errList)

		_, data, ok := sts.GetForHostId(11665)
		assert.True(t, ok)
		fmt.Printf("DATA:'%+v'\n", data)

		assert.Len(t, errList, 0)
	})*/
}
