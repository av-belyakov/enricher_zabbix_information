package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/netip"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/enricher_zabbix_information/internal/appstorage"
	"github.com/av-belyakov/enricher_zabbix_information/test/helpersfile"
)

func TestAppStorageStatistics(t *testing.T) {
	var (
		storageSize int
	)

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

	as, err := appstorage.New()
	if err != nil {
		log.Fatalln(err)
	}

	t.Run("Тест 0. Начало выполнения процесса", func(t *testing.T) {
		as.SetProcessRunning()
		dateStart, dateEnd := as.GetDateExecution()
		assert.Equal(t, dateStart.Year(), time.Now().Year())
		assert.Equal(t, dateEnd.Year(), 1)
	})

	t.Run("Тест 1. Добавление данных", func(t *testing.T) {
		for _, v := range examleData.Hosts {
			hostId, err := strconv.Atoi(v.HostId)
			assert.NoError(t, err)

			as.AddElement(appstorage.HostDetailedInformation{
				HostId:       hostId,
				OriginalHost: v.Host,
			})
		}

		listElement := as.GetList()
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

			_, data, ok := as.GetForHostId(requiredHostId)
			assert.True(t, ok)
			assert.True(t, data.OriginalHost == requiredHost)
		})
		t.Run("Тест 2.2. Поиск по OriginalHost (медленный поиск)", func(t *testing.T) {
			var (
				requiredHostId int    = 11695
				requiredHost   string = "rg.ru"
			)

			index, data, ok := as.GetForOriginalHost(requiredHost)
			assert.True(t, ok)
			assert.Greater(t, index, -1)
			assert.True(t, data.HostId == requiredHostId)
		})
	})

	t.Run("Тест 3. Удаление элемента", func(t *testing.T) {
		requiredHostId := 12891

		as.DeleteElement(requiredHostId)
		assert.Less(t, len(as.GetList()), storageSize)

		_, _, ok := as.GetForHostId(requiredHostId)
		assert.False(t, ok)
	})

	t.Run("Тест 4. Поиск элементов с ошибками", func(t *testing.T) {
		list := as.GetListErrors()
		assert.Len(t, list, 0)

		hostId := 123456789
		ipHost, err := netip.ParseAddr("65.33.110.3")
		assert.NoError(t, err)
		as.AddElement(appstorage.HostDetailedInformation{
			Ips:          []netip.Addr{ipHost},
			HostId:       hostId,
			OriginalHost: "test.ru/anything&name=aa",
			DomainName:   "test.ru",
			Error:        errors.New("new test error"),
		})

		_, _, ok := as.GetForHostId(hostId)
		assert.True(t, ok)

		list = as.GetListErrors()
		assert.Len(t, list, 1)
	})

	t.Run("Тест 5. Модификация данных в элементе хранилища", func(t *testing.T) {
		hostId := 323671
		orgHost := "example-domain.ru/anything&name=aa"
		domainName := "example-domain.ru"

		as.AddElement(appstorage.HostDetailedInformation{
			HostId:       hostId,
			OriginalHost: orgHost,
		})

		ipHost1, err := netip.ParseAddr("101.34.78.63")
		assert.NoError(t, err)
		ipHost2, err := netip.ParseAddr("78.100.0.64")
		assert.NoError(t, err)
		ipHost3, err := netip.ParseAddr("18.6.53.36")
		assert.NoError(t, err)
		ipHost4, err := netip.ParseAddr("218.26.5.132")
		assert.NoError(t, err)

		t.Run("Тест 5.1. Изменение или добавление доменного имени", func(t *testing.T) {
			assert.NoError(t, as.SetDomainName(hostId, domainName))
		})
		t.Run("Тест 5.2. Изменение или добавление ip адресов", func(t *testing.T) {
			assert.NoError(t, as.SetIps(hostId, ipHost1))
			assert.NoError(t, as.SetIps(hostId, ipHost2, ipHost3, ipHost4))
		})
		t.Run("Тест 5.3. Изменение или добавление ошибки", func(t *testing.T) {
			assert.NoError(t, as.SetError(hostId, errors.New("new test error")))
		})
		t.Run("Тест 5.4. Валидация добавленных данных", func(t *testing.T) {
			_, data, ok := as.GetForHostId(hostId)
			assert.True(t, ok)
			assert.Equal(t, data.OriginalHost, orgHost)
			assert.Equal(t, data.DomainName, domainName)
			assert.Len(t, data.Ips, 4)
			assert.NotNil(t, data.Error)
		})
	})

	t.Run("Тест 6. Установка статуса процесса выполнения в 'false'", func(t *testing.T) {
		as.SetProcessNotRunning()
		dateStart, dateEnd := as.GetDateExecution()
		assert.Equal(t, dateStart.Year(), time.Now().Year())
		assert.Equal(t, dateEnd.Year(), time.Now().Year())
	})

	t.Run("Тест 7. Очистка всего хранилища", func(t *testing.T) {
		as.DeleteAll()
		assert.False(t, as.GetStatusProcessRunning())
		assert.Len(t, as.GetList(), 0)

		dateStart, dateEnd := as.GetDateExecution()
		assert.Equal(t, dateStart.Year(), 1)
		assert.Equal(t, dateEnd.Year(), 1)
	})

	t.Run("Тест 8. Конкурентный доступ к хранилищу", func(t *testing.T) {
		testList := map[string]struct {
			ipaddr       []netip.Addr
			originalHost string
		}{
			"example-111-domain.ru": {
				ipaddr: []netip.Addr{
					netip.MustParseAddr("101.34.78.63"),
					netip.MustParseAddr("211.48.71.162"),
					netip.MustParseAddr("11.13.178.45"),
				},
				originalHost: "example-111-domain.ru/anything&name=aa",
			},
			"example-222-domain.ru": {
				ipaddr: []netip.Addr{
					netip.MustParseAddr("121.13.238.23"),
					netip.MustParseAddr("212.148.170.126"),
					netip.MustParseAddr("106.183.181.245"),
				},
				originalHost: "example-222-domain.ru/anything&name=aa",
			},
			"example-333-domain.ru": {
				ipaddr: []netip.Addr{
					netip.MustParseAddr("151.14.98.6"),
					netip.MustParseAddr("11.40.1.12"),
					netip.MustParseAddr("101.113.18.5"),
				},
				originalHost: "example-333-domain.ru/anything&name=aa",
			},
			"example-444-domain.ru": {
				ipaddr: []netip.Addr{
					netip.MustParseAddr("21.13.38.3"),
					netip.MustParseAddr("21.158.171.16"),
					netip.MustParseAddr("16.183.181.25"),
				},
				originalHost: "example-444-domain.ru/anything&name=aa",
			},
			"example-555-domain.ru": {
				ipaddr: []netip.Addr{
					netip.MustParseAddr("151.14.98.6"),
					netip.MustParseAddr("1.4.165.182"),
					netip.MustParseAddr("161.13.198.15"),
				},
				originalHost: "example-555-domain.ru/anything&name=aa",
			},
			"example-666-domain.ru": {
				ipaddr: []netip.Addr{
					netip.MustParseAddr("222.133.86.3"),
					netip.MustParseAddr("91.158.19.116"),
					netip.MustParseAddr("56.187.183.125"),
				},
				originalHost: "example-666-domain.ru/anything&name=aa",
			},
		}

		t.Run("Тест 8.1. Добавление данных", func(t *testing.T) {
			var (
				wg  sync.WaitGroup
				num atomic.Int32
			)

			for k, v := range testList {
				num.Store(num.Load() + 1)

				k := k
				v := v
				n := num.Load()

				wg.Go(func() {
					as.AddElement(appstorage.HostDetailedInformation{
						HostId:       int(n),
						DomainName:   k,
						OriginalHost: v.originalHost,
						Ips:          v.ipaddr,
					})
				})
			}
			wg.Wait()

			assert.Equal(t, len(as.GetHosts()), len(testList))
		})

		t.Run("Тест 8.2. Чтение данных", func(t *testing.T) {
			domainNames := []string{}
			for k := range testList {
				domainNames = append(domainNames, k)
			}

			var (
				wg  sync.WaitGroup
				num atomic.Int32
			)

			for _, v := range domainNames {
				wg.Go(func() {
					if index, hostInfo, ok := as.GetForDomainName(v); ok {
						num.Store(num.Load() + 1)

						fmt.Printf(
							"index:'%d', hostId:'%d', OriginalHost:'%s', Ips:'%v'\n",
							index,
							hostInfo.HostId,
							hostInfo.OriginalHost,
							hostInfo.Ips,
						)
					}
				})
			}
			wg.Wait()

			assert.Equal(t, int(num.Load()), len(testList))
		})
	})

	//t.Run("Тест . ", func(t *testing.T) {})
}
