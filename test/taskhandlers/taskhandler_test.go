package applicationtamplateexample_test

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"testing"

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

// при запуске теста можно указать флаг -status-prod
// в тесте дополнительные флаги указываются через --,
// например -- -status-prod
// тогда тест будет запускатся с реальными данными авторизации
func TestTaskHandler(t *testing.T) {
	var (
		storageTemp *storage.ShortTermStorage
	)

	os.Setenv("GO_ENRICHERZI_MAIN", "test")

	appStatus := "test"
	for _, arg := range os.Args {
		if arg == "-status-prod" {
			appStatus = "prod"
		}
	}

	fmt.Println("appStatus:", appStatus)

	if appStatus == "prod" {
		if err := godotenv.Load("../../.env"); err != nil {
			t.Fatal(err)
		}
	}

	if err := godotenv.Load("../filesfortest/.env"); err != nil {
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
			listIp          []string
			listGroupsId    []string
			hostGroupList   *connectionjsonrpc.ResponseHostGroupList
			hostList        *connectionjsonrpc.ResponseHostList
			shortPrefixList netboxapi.ShortPrefixList

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
		})
		t.Run("Тест 2.3. Получаем список id хостов относящихся к группам вебсайтов мониторинга", func(t *testing.T) {
			listGroupsId, err = taskhandlers.GetListIdsWebsitesGroupMonitoring(constants.App_Dictionary_Path, hostGroupList.Result)
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

			for msg := range chInfo {
				err = storageTemp.SetIsProcessed(msg.HostId)
				assert.NoError(t, err)

				err = storageTemp.SetDomainName(msg.HostId, msg.DomainName)
				assert.NoError(t, err)

				err = storageTemp.SetIps(msg.HostId, msg.Ips[0], msg.Ips...)
				assert.NoError(t, err)
			}
		})
		t.Run("Тест 2.7. Получаем префиксы из Netbox", func(t *testing.T) {
			shortPrefixList = taskhandlers.GetNetboxPrefixes(ctx, netboxClient, logging)
			assert.Greater(t, shortPrefixList.Count, 0)
		})
		t.Run("Тест 2.8. Выполняем поиск ip адресов в префиксах полученных от Netbox", func(t *testing.T) {
			maxCountGoroutines := 10
			chQueue := make(chan storage.HostDetailedInformation, maxCountGoroutines)

			//неправильная логика ограничения количества горутин

			go func() {
				for item := range chQueue {
					go func(v storage.HostDetailedInformation) {
						for msg := range shortPrefixList.SearchIps(v.Ips) {
							storageTemp.SetNetboxHostId(v.HostId, msg.Id)
							storageTemp.SetSensorId(v.HostId, msg.SensorId)

							if msg.Status != "active" {
								continue
							}

							err = storageTemp.SetIsActive(v.HostId)
							assert.NoError(t, err)
						}
					}(item)
				}
			}()

			for _, v := range storageTemp.GetList() {
				chQueue <- v
			}

			close(chQueue)
		})
	})

	t.Cleanup(func() {
		os.Unsetenv("GO_ENRICHERZI_MAIN")
		ctxCancel()
	})
}
