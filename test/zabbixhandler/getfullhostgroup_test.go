package zabbixhandler

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/test/helpers"
	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"
)

func TestGetFullHostGroup(t *testing.T) {
	var (
		res           []byte
		hostGroupList *connectionjsonrpc.ResponseHostGroupList

		errMsg *connectionjsonrpc.ResponseError
		err    error
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

	t.Run("Тест 1. Авторизация и аутентификация в Zabbix", func(t *testing.T) {
		assert.NoError(t, zabbixConn.AuthorizationStart(ctx))
	})

	t.Run("Тест 2.1. Получаем полный список групп хостов Zabbix", func(t *testing.T) {
		res, err = zabbixConn.GetFullHostGroupList(ctx)
		assert.NoError(t, err)
	})
	t.Run("Тест 2.2. Преобразуем список групп хостов Zabbix из бинарного вида в JSON", func(t *testing.T) {
		hostGroupList, errMsg, err = connectionjsonrpc.NewResponseGetHostGroupList().Get(res)
		assert.NoError(t, err)
		assert.Equal(t, errMsg.Error.Code, 0)
		assert.NotEmpty(t, hostGroupList.Result)

		for k, v := range hostGroupList.Result {
			if k == 10 {
				break
			}

			fmt.Printf("%d.\n\t%+v\n", k+1, v)
		}
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
