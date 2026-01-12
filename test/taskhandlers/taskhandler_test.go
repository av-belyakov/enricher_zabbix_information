package applicationtamplateexample_test

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
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
	os.Setenv("GO_ENRICHERZI_MAIN", "test")

	appStatus := "test"
	flag.StringVar(&appStatus, "status-prod", "", "start test at prod status")
	flag.Parse()

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

	taskHandlerSettings := taskhandlers.NewSettings(
		zabbixConn,
		&apiserver.InformationServer{},
		storage.NewShortTermStorage(),
		logging,
	)

	taskHandler := taskHandlerSettings.Init(ctx)

	t.Run("Тест 1. Авторизация и аутентификация в Zabbix", func(t *testing.T) {
		assert.NoError(t, zabbixConn.AuthorizationStart(ctx))
	})

	t.Run("Тест 2. Проверка обработчика задач", func(t *testing.T) {
		assert.NoError(t, taskHandler.SimpleTaskHandler())
	})

	t.Cleanup(func() {
		os.Unsetenv("GO_ENRICHERZI_MAIN")
		ctxCancel()
	})
}
