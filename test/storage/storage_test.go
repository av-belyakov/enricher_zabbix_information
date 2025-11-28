package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/netip"
	"os"
	"strconv"
	"testing"

	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestStorage(t *testing.T) {
	var (
		storageSize int
	)

	b, err := os.ReadFile("./exampledata.json")
	if err != nil {
		log.Fatalln(err)
	}

	examleData := TypeExampleData{}
	if err := json.Unmarshal(b, &examleData); err != nil {
		log.Fatalln(err)
	}
	if len(examleData.Hosts) == 0 {
		log.Fatalln(errors.New("the structure 'TypeExampleData' should not be empty"))
	}

	sts := storage.NewShortTermStorage()

	t.Run("Тест 1. Добавление данных", func(t *testing.T) {
		for _, v := range examleData.Hosts {
			hostId, err := strconv.Atoi(v.HostId)
			assert.NoError(t, err)

			sts.Add(storage.HostDetailedInformation{
				HostId:       hostId,
				OriginalHost: v.Host,
			})
		}

		listElement := sts.GetList()
		storageSize = len(listElement)

		fmt.Println("It was added", storageSize, " elements")

		assert.Greater(t, storageSize, 0)
	})

	t.Run("Тест 2. Поиск информации в хранилище", func(t *testing.T) {
		t.Run("Тест 2.1. Поиск по HostId (быстрый поиск)", func(t *testing.T) {
			var (
				requiredHostId int    = 11682
				requiredHost   string = "feedback.sk.ru"
			)

			data, ok := sts.GetForHostId(requiredHostId)
			assert.True(t, ok)
			assert.True(t, data.OriginalHost == requiredHost)
		})
		t.Run("Тест 2.2. Поиск по OriginalHost (медленный поиск)", func(t *testing.T) {
			var (
				requiredHostId int    = 11695
				requiredHost   string = "rg.ru"
			)

			index, data, ok := sts.GetForOriginalHost(requiredHost)
			assert.True(t, ok)
			assert.Greater(t, index, -1)
			assert.True(t, data.HostId == requiredHostId)
		})
	})

	t.Run("Тест 3. Удаление элемента", func(t *testing.T) {
		requiredHostId := 12891

		sts.DeleteElement(requiredHostId)
		assert.Less(t, len(sts.GetList()), storageSize)

		_, ok := sts.GetForHostId(requiredHostId)
		assert.False(t, ok)
	})

	t.Run("Тест 4. Поиск элементов с ошибками", func(t *testing.T) {
		list := sts.GetListErrors()
		assert.Len(t, list, 0)

		hostId := 123456789
		ipHost, err := netip.ParseAddr("65.33.110.3")
		assert.NoError(t, err)
		sts.Add(storage.HostDetailedInformation{
			Ip:           []netip.Addr{ipHost},
			HostId:       hostId,
			OriginalHost: "test.ru",
			Errors:       errors.New("new test error"),
		})

		_, ok := sts.GetForHostId(hostId)
		assert.True(t, ok)

		list = sts.GetListErrors()
		assert.Len(t, list, 1)
	})

	t.Run("Тест 5. Очистка всего хранилища", func(t *testing.T) {
		sts.DeleteAll()
		assert.Len(t, sts.GetList(), 0)
	})

	//t.Run("Тест . ", func(t *testing.T) {})
}

type TypeExampleData struct {
	Hosts []struct {
		Name   string `json:"name"`
		Host   string `json:"host"`
		HostId string `json:"host_id"`
	}
}
