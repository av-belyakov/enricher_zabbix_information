package applicationtamplateexample_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
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

//
// Для этого теста надо ОБЯЗАТЕЛЬНО развернуть тестовый Zabbix.
//

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

	testZabbixConn, err := zabbixConnectForTesting()
	if err != nil {
		t.Fatal(err)
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
		assert.NoError(t, testZabbixConn.AuthorizationStart(ctx))
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
			res, err = testZabbixConn.GetFullHostGroupList(ctx)
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
			res, err = testZabbixConn.GetHostList(ctx, listGroupsId...)
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

	t.Run("Тест 3. Добавление или обновление тегов в тестовом Zabbix", func(t *testing.T) {
		listHostWithSensorId := storageTemp.GetHostsWithSensorId()
		assert.Greater(t, len(listHostWithSensorId), 0)

		var num int
		fmt.Println("Insert tags to Zabbix")
		for _, v := range listHostWithSensorId {
			if num == 10 {
				break
			}

			hostId := fmt.Sprint(v.HostId)
			fmt.Println("___ fmt.Sprint(v.HostId):", hostId)

			res, err := testZabbixConn.GetHostTags(ctx, hostId)
			assert.NoError(t, err)

			fmt.Printf("Response GetHostTags: '%s'\n", res)
			num++

			/*

				1. Есть ошибки the host with id '11617' was not found.
				Похоже для поиска я использую данные с одного Zabbix (продуктового),
				а изменить пытаюсь данные в тестовом Zabbix. Надо использовать только
				тестовый Zabbix. Тем более данные в нем почти такие же как и в продуктовом.

				!Вроде уже нет ошибок, но всё равно теги похоже не обнавляются!


				2. Есть гонка данных. Надо смотреть временное хранилище. В тестоах временного
				хранилища всё нормально. Однако, в тестах не предусмотрен параллельный доступ.

			*/

			/*

				!!! Раскомментировать для изменения тегов в тестовом Zabbix !!!

					var sensorsId string
					countSensorsId := len(v.SensorsId)
					if countSensorsId == 0 {
						continue
					} else if countSensorsId == 1 {
						sensorsId = v.SensorsId[0]
					} else {
						sensorsId = strings.Join(v.SensorsId, ",")
					}

					res, err := testZabbixConn.UpdateHostParameterTags(
						ctx,
						fmt.Sprint(v.HostId),
						connectionjsonrpc.Tags{
							Tag: []connectionjsonrpc.Tag{{Tag: "СОА-ТЕСТ", Value: sensorsId}},
						},
					)
					assert.NoError(t, err)

					fmt.Printf("Response UpdateHostParameterTags: '%s'\n", string(res))
			*/
		}
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
