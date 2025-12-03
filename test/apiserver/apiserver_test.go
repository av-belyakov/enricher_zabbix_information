package apiserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appversion"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/av-belyakov/enricher_zabbix_information/test/helpers"
	"github.com/stretchr/testify/assert"
)

func TestApiServer(t *testing.T) {
	const (
		host = "localhost"
		port = 8989
	)

	var (
		res *http.Response
		err error
	)

	logging := helpers.NewLoggingForTest()

	ctx, ctxCancel := context.WithCancel(t.Context())
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

	t.Run("Тест 1. Запуск API сервера", func(t *testing.T) {
		version, err := appversion.GetVersion()
		assert.NoError(t, err)

		api, err := apiserver.New(
			logging,
			storageTemp,
			apiserver.WithHost(host),
			apiserver.WithPort(port),
			apiserver.WithVersion(version),
		)
		assert.NoError(t, err)

		api.Start(ctx)

		time.Sleep(time.Second * 1)
	})

	t.Run("Тест 2. Обращение к странице с задачами", func(t *testing.T) {
		res, err = http.Get("http://" + net.JoinHostPort(host, fmt.Sprint(port)) + "/tasks")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, res.StatusCode, http.StatusOK)

		b, err := io.ReadAll(res.Body)
		res.Body.Close()
		assert.NoError(t, err)

		fmt.Println("Web server response:", string(b))
	})

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
