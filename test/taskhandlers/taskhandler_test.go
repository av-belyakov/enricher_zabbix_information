package applicationtamplateexample_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"testing/synctest"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appstorage"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/dnsresolver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/netboxapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/internal/taskhandlers"
	"github.com/av-belyakov/enricher_zabbix_information/test/helpers"
)

//
// Для этого теста надо ОБЯЗАТЕЛЬНО развернуть тестовый Zabbix.
//
// Netbox используется продуктовый (во всяком случае пока)
//

func TestTaskHandler(t *testing.T) {
	var (
		storageTemp *appstorage.SharedAppStorage
	)

	f, err := os.Create("cpupprof.out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err = pprof.StartCPUProfile(f); err != nil {
		t.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// это нужно для доступа к Netbox
	os.Setenv("GO_ENRICHERZI_MAIN", "production")

	// читаем продуктовые перменные окружения только что бы получить
	// доступ к продуктовому Netbox
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

	// соединение с Zabbix
	testZabbixConn, err := zabbixConnectForTesting()
	if err != nil {
		t.Fatal(err)
	}

	// соединение с Netbox
	netboxClient, err := netboxapi.New(
		conf.NetBox.Host,
		conf.NetBox.Port,
		conf.AuthenticationData.NetBoxToken,
	)
	if err != nil {
		log.Fatalf("error initializing the Netbox client: %v", err)
	}

	storageTemp, err = appstorage.New()
	if err != nil {
		log.Fatalln(err)
	}

	t.Run("Тест 1. Авторизация и аутентификация в Zabbix", func(t *testing.T) {
		assert.NoError(t, testZabbixConn.AuthorizationStart(ctx))
	})

	t.Run("Тест 2. Проверка шагов обработчика задач", func(t *testing.T) {
		var (
			res             []byte
			listGroupsId    []string
			hostGroupList   *connectionjsonrpc.ResponseHostGroupList
			hostList        *connectionjsonrpc.ResponseHostList
			shortPrefixList netboxapi.ShortPrefixList = netboxapi.ShortPrefixList{}

			errMsg *connectionjsonrpc.ResponseError
			err    error
		)

		t.Run("Тест 2.1. Получаем полный список групп хостов Zabbix", func(t *testing.T) {
			res, err = testZabbixConn.GetFullHostGroupList(ctx)
			assert.NoError(t, err)
		})
		t.Run("Тест 2.2. Преобразуем список групп хостов Zabbix из бинарного вида в структуру", func(t *testing.T) {
			hostGroupList, errMsg, err = connectionjsonrpc.NewResponseGetHostGroupList().Get(res)
			assert.NoError(t, err)
			assert.Equal(t, errMsg.Error.Code, 0)
			assert.NotEmpty(t, hostGroupList.Result)

			// количество групп хостов в Zabbix
			storageTemp.SetCountZabbixHostsGroup(len(hostGroupList.Result))

			fmt.Println("Count host group list:", len(hostGroupList.Result))
		})
		t.Run("Тест 2.3. Получаем список id хостов относящихся к группам веб-сайтов мониторинга", func(t *testing.T) {
			listGroupsId, err = taskhandlers.GetListIdsWebsitesGroupMonitoring(constants.App_Dictionary_Path, hostGroupList.Result)
			assert.NoError(t, err)

			// количество групп хостов относящихся к веб-сайтам мониторинга
			storageTemp.SetCountMonitoringHostsGroup(len(listGroupsId))

			fmt.Println("Count list groups id:", len(listGroupsId))
		})
		t.Run("Тест 2.4. Получаем список хостов которые есть в словарях, если словари не пусты, или все хосты", func(t *testing.T) {
			res, err = testZabbixConn.GetHostList(ctx, listGroupsId...)
			assert.NoError(t, err)
		})
		t.Run("Тест 2.5. Преобразуем список хостов Zabbix из бинарного вида в структуру", func(t *testing.T) {
			hostList, errMsg, err = connectionjsonrpc.NewResponseGetHostList().Get(res)
			assert.NoError(t, err)
			assert.Equal(t, errMsg.Error.Code, 0)
			assert.NotEmpty(t, hostList.Result)

			// количество хостов по которым осуществляется мониторинг
			storageTemp.SetCountMonitoringHosts(len(hostList.Result))

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
					storageTemp.AddElement(appstorage.HostDetailedInformation{
						HostId:       hostId,
						OriginalHost: host.Host,
					})
				}
			}

			// инициализируем поиск ip адресов через DNS resolver
			dnsRes, err := dnsresolver.New(dnsresolver.WithTimeout(10))
			assert.NoError(t, err)

			// запускаем поиск через DNS resolver
			chInfo, err := dnsRes.Run(ctx, storageTemp.GetHosts())
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
			chunPrefixInfo, count, err := taskhandlers.NetboxPrefixes(ctx, netboxClient, logging)
			assert.NoError(t, err)
			assert.Greater(t, count, 0)

			// количество найденных префиксов в Netbox
			storageTemp.SetCountNetboxPrefixes(count)

			fmt.Println("Count short prefix list:", count)

			for prefixInfo := range chunPrefixInfo {
				shortPrefixList = append(shortPrefixList, prefixInfo...)

				//отправка в apiserver
			}

		})
		t.Run("Тест 2.8. Выполняем поиск ip адресов в префиксах полученных от Netbox", func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				taskhandlers.SearchIpaddrToPrefixesNetbox(3, storageTemp, shortPrefixList, logging)
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

	t.Run("Тест 3. Добавляем или обновляем теги в тестовом Zabbix", func(t *testing.T) {
		listHostWithSensorId := storageTemp.GetHostsWithSensorId()
		assert.Greater(t, len(listHostWithSensorId), 0)

		var num int
		fmt.Println("Insert tags to Zabbix. Print first 10 items.")
		for _, v := range listHostWithSensorId {
			if num <= 10 {
				hostId := fmt.Sprint(v.HostId)
				fmt.Println("___ fmt.Sprint(v.HostId):", hostId)
			}

			num++

			//!!! Раскомментировать для изменения тегов в тестовом Zabbix !!!
			//-------------
			var sensorsId string
			if len(v.SensorsId) == 0 {
				continue
			} else if len(v.SensorsId) == 1 {
				sensorsId = v.SensorsId[0]
			} else {
				sensorsId = strings.Join(v.SensorsId, ",")
			}

			_, err := testZabbixConn.UpdateHostParameterTags(
				ctx,
				fmt.Sprint(v.HostId),
				connectionjsonrpc.Tags{
					Tag: []connectionjsonrpc.Tag{
						{Tag: "СОА-ТЕСТ", Value: sensorsId},
					},
				},
			)
			assert.NoError(t, err)

			//fmt.Printf("Response UpdateHostParameterTags: '%s'\n", string(b))
		}

		// количество обновленных хостов в Zabbix
		storageTemp.SetCountUpdatedZabbixHosts(num)
	})

	t.Cleanup(func() {
		os.Unsetenv("GO_ENRICHERZI_MAIN")

		os.Unsetenv("GO_ENRICHERZI_ZPASSWD")
		os.Unsetenv("GO_ENRICHERZI_NBTOKEN")
		os.Unsetenv("GO_ENRICHERZI_DBWLOGPASSWD")
		os.Unsetenv("GO_ENRICHERZI_APISERVERTOKEN")

		os.Unsetenv("GO_TESTZABBIX_HOST")
		os.Unsetenv("GO_TESTZABBIX_PORT")
		os.Unsetenv("GO_TESTZABBIX_USER")
		os.Unsetenv("GO_TESTZABBIX_PASSWD")

		ctxCancel()
	})
}

// zabbixConnectForTesting подключение к тестовому Zabbix
func zabbixConnectForTesting() (*connectionjsonrpc.ZabbixConnectionJsonRPC, error) {
	if err := godotenv.Load(".env.test"); err != nil {
		return nil, err
	}

	testZabbixHost := os.Getenv("GO_TESTZABBIX_HOST")
	if testZabbixHost == "" {
		return nil, errors.New("environment variable 'GO_TESTZABBIX_HOST' cannot be empty")
	}

	tmpPort := os.Getenv("GO_TESTZABBIX_PORT")
	if tmpPort == "" {
		return nil, errors.New("environment variable 'GO_TESTZABBIX_PORT' cannot be empty")
	}
	testZabbixPort, err := strconv.Atoi(tmpPort)
	if err != nil {
		return nil, err
	}

	testZabbixUser := os.Getenv("GO_TESTZABBIX_USER")
	if testZabbixUser == "" {
		return nil, errors.New("environment variable 'GO_TESTZABBIX_USER' cannot be empty")
	}

	testZabbixPasswd := os.Getenv("GO_TESTZABBIX_PASSWD")
	if testZabbixPasswd == "" {
		return nil, errors.New("environment variable 'GO_TESTZABBIX_PASSWD' cannot be empty")
	}

	fmt.Printf("zabbix config user:'%s', passwd:'%s'\n", testZabbixUser, testZabbixPasswd)

	conn, err := connectionjsonrpc.NewConnect(
		connectionjsonrpc.WithHost(testZabbixHost),
		connectionjsonrpc.WithPort(testZabbixPort),
		connectionjsonrpc.WithLogin(testZabbixUser),
		connectionjsonrpc.WithPasswd(testZabbixPasswd),
		connectionjsonrpc.WithConnectionTimeout(10),
	)
	if err != nil {
		return nil, fmt.Errorf("error zabbix connection: %v", err)
	}

	return conn, nil
}
