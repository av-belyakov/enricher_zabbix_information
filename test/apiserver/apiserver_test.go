package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appversion"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/av-belyakov/enricher_zabbix_information/test/helpers"
)

type ElementManuallyTask struct {
	Type     string                      `json:"type"`
	Settings ElementManuallyTaskSettings `json:"settings"`
}

type ElementManuallyTaskSettings struct {
	Command string `json:"command"`
	Error   string `json:"error"`
	Token   string `json:"token"`
}

func TestApiServer(t *testing.T) {
	const (
		host = "localhost"
		port = 8989
	)

	var (
		api *apiserver.InformationServer
		res *http.Response
		err error
	)

	logging := helpers.NewLoggingForTest()
	ctx, ctxCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	//ctx, ctxCancel := context.WithCancel(t.Context())

	go func() {
		for {
			select {
			case <-ctx.Done():
				ctxCancel()

				return

			case msg := <-logging.GetChan():
				fmt.Printf("Log message: type:'%s', message:'%s'\n", msg.GetType(), msg.GetMessage())
			}
		}
	}()

	storageTemp := storage.NewShortTermStorage()
	storageTemp.SetProcessRunning()
	fillInStorage(storageTemp)
	time.Sleep(time.Second * 2)
	storageTemp.SetProcessNotRunning()

	version, err := appversion.GetVersion()
	assert.NoError(t, err)

	//инициализируем api сервер
	api, err = apiserver.New(
		logging,
		storageTemp,
		apiserver.WithHost(host),
		apiserver.WithPort(port),
		apiserver.WithVersion(version),
	)
	if err != nil {
		log.Fatal(err)
	}

	//запускаем api сервер
	go api.Start(ctx)

	time.Sleep(time.Second * 1)

	//обработчик входящих от api сервера данных
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case b := <-api.GetChannelOutgoingData():
				fmt.Println("Outgoing data from API server:", string(b))
				elmManualTask := ElementManuallyTask{}

				err := json.Unmarshal(b, &elmManualTask)
				assert.NoError(t, err)

				if elmManualTask.Type != "manually_task" {
					continue
				}

				var strErr string
				if elmManualTask.Settings.Token != "my_some_token" {
					strErr = "token invalide"
				}

				b, err = json.Marshal(ElementManuallyTask{
					Type: "manually_task",
					Settings: ElementManuallyTaskSettings{
						Error: strErr,
					},
				})
				assert.NoError(t, err)

				api.SendData(b)
			}
		}
	}()

	//t.Run("Тест 1. Запуск API сервера", func(t *testing.T) {})

	t.Run("Тест 1. Обращение к странице с задачами", func(t *testing.T) {
		res, err = http.Get("http://" + net.JoinHostPort(host, fmt.Sprint(port)) + "/manually_task_starting")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, res.StatusCode, http.StatusOK)

		//b, err := io.ReadAll(res.Body)
		//assert.NoError(t, err)

		//fmt.Println("Web server response:", string(b))

		res.Body.Close()
	})

	t.Run("Тест 2. Передача сообщения на веб-страницу используя websocket", func(t *testing.T) {
		for {
			select {
			case <-ctx.Done():
				return

			case <-time.After(time.Second * 3):
				api.SendData(fmt.Appendf(nil, `{
					"type": "logs",
					"data": {
						"timestamp": "%s",
						"level": "WARNING",
						"message": "Warning test message"
					}
				}`, time.Now().Format(time.RFC3339)))

				time.Sleep(time.Second * 1)

				api.SendData(fmt.Appendf(nil, `{
					"type": "logs",
					"data": {
						"timestamp": "%s",
						"level": "INFO",
						"message": "Info test message"
					}
				}`, time.Now().Format(time.RFC3339)))

			}
		}
	})

	<-ctx.Done()

	t.Cleanup(func() {
		if res != nil {
			res.Body.Close()
		}

		ctxCancel()
	})
}

func fillInStorage(s *storage.ShortTermStorage) {
	testList := []datamodels.HostDetailedInformation{
		{
			HostId:       14421,
			OriginalHost: "1yar.tv",
		},
		{
			HostId:       14412,
			OriginalHost: "tfoms-rzn.ru",
		},
		{
			HostId:       14433,
			OriginalHost: "nrcki.ru",
		},
		{
			HostId:       14582,
			OriginalHost: "rosatom.ruindex.html",
			Error:        errors.New("rosatom.ruindex.html' was not found, learn more (lookup rosatom.ruindex.html on 127.0.0.53:53: no such host)"),
		},
		{
			HostId:       14521,
			OriginalHost: "www.kremlin.ru2",
			Error:        errors.New("www.kremlin.ru2' was not found, learn more (lookup www.kremlin.ru2 on 127.0.0.53:53: no such host)"),
		},
		{
			HostId:       14438,
			OriginalHost: "www.energia.ru",
		},
		{
			HostId:       14438,
			OriginalHost: "rptp.org",
		},
		{
			HostId:       14441,
			OriginalHost: "www.invest-in-voronezh.ruru",
			Error:        errors.New("www.invest-in-voronezh.ruru' was not found, learn more (lookup www.invest-in-voronezh.ruru on 127.0.0.53:53: no such host)"),
		},
	}

	for _, v := range testList {
		s.Add(v)
	}
}
