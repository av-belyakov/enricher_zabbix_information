package apiserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

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
