package applicationtamplateexample_test

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"testing/synctest"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/dnsresolver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/netboxapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/internal/taskhandlers"
	"github.com/av-belyakov/enricher_zabbix_information/test/helpers"
)

func TestTaskHandler(t *testing.T) {
	var (
		storageTemp *storage.ShortTermStorage
	)

	os.Setenv("GO_ENRICHERZI_MAIN", "production")

	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatal(err)
	}

	rootPath, err := supportingfunctions.GetRootPath(constants.Root_Dir)
	if err != nil {
		t.Fatalf("Не удалось получить корневую директорию: %v", err)
	}

	conf, err := confighandler.New(rootPath)
	if err != nil {
		t.Fatalf("Не удалось прочитать конфигурационный файл: %v", err)
	}

	logging := helpers.NewLoggingForTest()
	ctx, ctxCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case msg := <-logging.GetChan():
				fmt.Printf("Log message: type:'%s', message:'%s'\n", msg.GetType(), msg.GetMessage())
			}
		}
	}()

	fmt.Printf("zabbix config user:'%s', passwd:'%s'\n", conf.GetZabbix().User, conf.GetAuthenticationData().ZabbixPasswd)

	zabbixConn, err := connectionjsonrpc.NewConnect(
		connectionjsonrpc.WithTLS(),
		connectionjsonrpc.WithInsecureSkipVerify(),
		connectionjsonrpc.WithHost(conf.GetZabbix().Host),
		connectionjsonrpc.WithPort(conf.GetZabbix().Port),
		connectionjsonrpc.WithLogin(conf.GetZabbix().User),
		connectionjsonrpc.WithPasswd(conf.GetAuthenticationData().ZabbixPasswd),
		connectionjsonrpc.WithConnectionTimeout(cmp.Or(conf.GetZabbix().Timeout, 10)),
	)
	if err != nil {
		t.Fatalf("error zabbix connection: %v", err)
	}

	netboxClient, err := netboxapi.New(
		conf.NetBox.Host,
		conf.NetBox.Port,
		conf.AuthenticationData.NetBoxToken,
	)
	if err != nil {
		log.Fatalf("error initializing the Netbox client: %v", err)
	}

	storageTemp = storage.NewShortTermStorage()

	t.Run("Тест 1. Авторизация и аутентификация в Zabbix", func(t *testing.T) {
		assert.NoError(t, zabbixConn.AuthorizationStart(ctx))
	})

	t.Run("Тест 2. Проверка шагов обработчика задач", func(t *testing.T) {
		var (
			res             []byte
			listGroupsId    []string
			hostGroupList   *connectionjsonrpc.ResponseHostGroupList
			hostList        *connectionjsonrpc.ResponseHostList
			shortPrefixList *netboxapi.ShortPrefixList

			errMsg *connectionjsonrpc.ResponseError
			err    error
		)

		t.Run("Тест 2.1. Получаем полный список групп хостов Zabbix", func(t *testing.T) {
			res, err = zabbixConn.GetFullHostGroupList(ctx)
			assert.NoError(t, err)
		})
		t.Run("Тест 2.2. Преобразуем список групп хостов Zabbix из бинарного вида в JSON", func(t *testing.T) {
			hostGroupList, errMsg, err = connectionjsonrpc.NewResponseGetHostGroupList().Get(res)
			assert.NoError(t, err)
			assert.Equal(t, errMsg.Error.Code, 0)
			assert.NotEmpty(t, hostGroupList.Result)

			fmt.Println("Count host group list:", len(hostGroupList.Result))
		})
		t.Run("Тест 2.3. Получаем список id хостов относящихся к группам веб-сайтов мониторинга", func(t *testing.T) {
			listGroupsId, err = taskhandlers.GetListIdsWebsitesGroupMonitoring(constants.App_Dictionary_Path, hostGroupList.Result)
			assert.NoError(t, err)

			fmt.Println("Count list groups id:", len(listGroupsId))
		})
		t.Run("Тест 2.4. Получаем список хостов которые есть в словарях, если словари не пусты, или все хосты", func(t *testing.T) {
			res, err = zabbixConn.GetHostList(ctx, listGroupsId...)
			assert.NoError(t, err)
		})
		t.Run("Тест 2.5. Преобразуем список хостов Zabbix из бинарного вида в JSON", func(t *testing.T) {
			hostList, errMsg, err = connectionjsonrpc.NewResponseGetHostList().Get(res)
			assert.NoError(t, err)
			assert.Equal(t, errMsg.Error.Code, 0)
			assert.NotEmpty(t, hostList.Result)

			fmt.Println("Count host list:", len(hostList.Result))
			//for k, v := range hostList.Result {
			//	fmt.Printf("%d. name:'%s', host id:'%s', host:'%s'\n", k+1, v.Name, v.HostId, v.Host)
			//}
		})
		t.Run("Тест 2.6. Инициализируем поиск ip адресов через DNS resolver", func(t *testing.T) {
			//очищаем хранилище от предыдущих данных (что бы не смешивать старые и новые данные)
			storageTemp.DeleteAll()
			// устанавливаем дату начала выполнения задачи
			storageTemp.SetStartDateExecution()

			// заполняем хранилище данными о хостах
			for _, host := range hostList.Result {
				if hostId, err := strconv.Atoi(host.HostId); err == nil {
					storageTemp.Add(storage.HostDetailedInformation{
						HostId:       hostId,
						OriginalHost: host.Host,
					})
				}
			}

			// инициализируем поиск ip адресов через DNS resolver
			dnsRes, err := dnsresolver.New(
				storageTemp,
				dnsresolver.WithTimeout(10),
			)
			assert.NoError(t, err)

			// запускаем поиск через DNS resolver
			chInfo, err := dnsRes.Run(ctx)
			assert.NoError(t, err)

			//var num int
			for msg := range chInfo {
				//num++
				//fmt.Printf("%d. Chun. Original host:'%s', ips:'%+v', error:'%v'\n", num, msg.OriginalHost, msg.Ips, msg.Error)

				err = storageTemp.SetDomainName(msg.HostId, msg.DomainName)
				assert.NoError(t, err)

				err = storageTemp.SetIps(msg.HostId, msg.Ips...)
				assert.NoError(t, err)
			}
		})
		t.Run("Тест 2.7. Получаем префиксы из Netbox", func(t *testing.T) {
			shortPrefixList = taskhandlers.GetNetboxPrefixes(ctx, netboxClient, logging)
			assert.Greater(t, shortPrefixList.Count, 0)

			fmt.Println("Count short prefix list:", shortPrefixList.Count)
		})
		t.Run("Тест 2.8. Выполняем поиск ip адресов в префиксах полученных от Netbox", func(t *testing.T) {
			maxCountGoroutines := 3 // runtime.NumCPU()
			chQueue := make(chan storage.HostDetailedInformation, maxCountGoroutines)

			//fmt.Println("Count CPUs:", maxCountGoroutines)

			synctest.Test(t, func(t *testing.T) {
				var wg sync.WaitGroup
				ids := []int{
					11585,
					14314,
					14315,
					14316,
					14317,
				}

				// создаём обработчики задач
				//for i := range maxCountGoroutines {
				for range maxCountGoroutines {
					wg.Go(func() {
						for hostInfo := range chQueue {
							//if slices.Contains(ids, hostInfo.HostId) {
							//	fmt.Printf("host id: '%d', all info: '%+v'\n", hostInfo.HostId, hostInfo)
							//}
							//fmt.Printf("Goroutine %d, host id: '%d', addr: '%+v'\n", i, hostInfo.HostId, hostInfo.Ips)

							for msgList := range shortPrefixList.SearchIps(hostInfo.Ips) {
								//-------- debug --------
								if slices.Contains(ids, hostInfo.HostId) {
									fmt.Printf("host id: '%d', all info: '%+v', ipaddr info: '%+v'\n", hostInfo.HostId, hostInfo, msgList)
								}
								//-----------------------

								err = storageTemp.SetIsProcessed(hostInfo.HostId)
								assert.NoError(t, err)

								for _, msg := range msgList {
									if msg.Status != "active" {
										continue
									}

									err = storageTemp.SetIsActive(hostInfo.HostId)
									assert.NoError(t, err)

									err = storageTemp.SetSensorId(hostInfo.HostId, msg.SensorId)
									assert.NoError(t, err)

									err = storageTemp.SetNetboxHostId(hostInfo.HostId, msg.Id)
									assert.NoError(t, err)
								}
							}
						}
					})
				}

				for _, v := range storageTemp.GetList() {
					chQueue <- v
				}

				close(chQueue)

				wg.Wait()
			})

			listHostWithSensorId := storageTemp.GetHostsWithSensorId()
			assert.Greater(t, len(listHostWithSensorId), 0)

			fmt.Println("Hosts with sensor id:")
			for k, v := range listHostWithSensorId {
				fmt.Printf(
					"%d. domain name:'%s', host id:%d, ips:'%v', sensors id:'%s', netbox hosts id:'%v', IsProcessed:'%t', error:'%v'\n",
					k+1,
					v.DomainName,
					v.HostId,
					v.Ips,
					v.SensorsId,
					v.NetboxHostsId,
					v.IsProcessed,
					v.Error,
				)
			}
		})
	})

	t.Cleanup(func() {
		os.Unsetenv("GO_ENRICHERZI_MAIN")

		os.Unsetenv("GO_ENRICHERZI_ZPASSWD")
		os.Unsetenv("GO_ENRICHERZI_NBTOKEN")
		os.Unsetenv("GO_ENRICHERZI_DBWLOGPASSWD")
		os.Unsetenv("GO_ENRICHERZI_APISERVERTOKEN")

		ctxCancel()
	})
}
