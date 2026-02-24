package main

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/av-belyakov/simplelogger"
	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appstorage"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appversion"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/elasticsearchapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/logginghandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/netboxapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/schedulehandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/internal/taskhandlers"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

func app(ctx context.Context) {
	var (
		nameRegionalObject string
		zabbixConn         *connectionjsonrpc.ZabbixConnectionJsonRPC
		err                error
	)

	if os.Getenv("GO_"+constants.App_Environment_Name+"_MAIN") == "development" {
		nameRegionalObject = fmt.Sprintf("%s-dev", appname.GetName())
	} else {
		nameRegionalObject = appname.GetName()
	}

	rootPath, err := supportingfunctions.GetRootPath(constants.Root_Dir)
	if err != nil {
		log.Fatalf("error, it is impossible to form root path (%s)", err.Error())
	}

	versionApp, err := appversion.GetVersion()
	if err != nil {
		versionApp = "v0.0.0"
	}

	// ****************************************************************************
	// *********** инициализируем модуль чтения конфигурационного файла ***********
	cfg, err := confighandler.New(rootPath)
	if err != nil {
		log.Fatalf("error module 'confighandler': %v", err)
	}

	// ****************************************************************************
	// ********************* инициализация модуля логирования *********************
	var listLog []simplelogger.OptionsManager
	for _, v := range cfg.GetListLogs() {
		listLog = append(listLog, v)
	}
	opts := simplelogger.CreateOptions(listLog...)
	simpleLogger, err := simplelogger.NewSimpleLogger(ctx, constants.Root_Dir, opts)
	if err != nil {
		log.Fatalf("error module 'simplelogger': %v", err)
	}

	//*********************************************************************************
	//********** инициализация модуля взаимодействия с БД для передачи логов **********
	cfgDB := cfg.GetLogDB()
	if esc, err := elasticsearchapi.NewElasticsearchConnect(elasticsearchapi.Settings{
		Port:               cfgDB.Port,
		Host:               cfgDB.Host,
		User:               cfgDB.User,
		Passwd:             cfg.GetAuthenticationData().WriteLogBDPasswd,
		IndexDB:            cfgDB.StorageNameDB,
		NameRegionalObject: nameRegionalObject,
	}); err != nil {
		_ = simpleLogger.Write("error", wrappers.WrapperError(err).Error())
	} else {
		//подключение логирования в БД
		simpleLogger.SetDataBaseInteraction(esc)
	}

	//*********************************************************************************
	//***************** инициализация модуля-обёртки соединения с Zabbix **************
	if cfg.GetZabbix().UseTLS {
		zabbixConn, err = connectionjsonrpc.NewConnect(
			connectionjsonrpc.WithTLS(),
			connectionjsonrpc.WithInsecureSkipVerify(),
			connectionjsonrpc.WithHost(cfg.GetZabbix().Host),
			connectionjsonrpc.WithPort(cfg.GetZabbix().Port),
			connectionjsonrpc.WithLogin(cfg.GetZabbix().User),
			connectionjsonrpc.WithPasswd(cfg.GetAuthenticationData().ZabbixPasswd),
			connectionjsonrpc.WithConnectionTimeout(cmp.Or(cfg.GetZabbix().Timeout, 10)),
		)
	} else {
		zabbixConn, err = connectionjsonrpc.NewConnect(
			connectionjsonrpc.WithHost(cfg.GetZabbix().Host),
			connectionjsonrpc.WithPort(cfg.GetZabbix().Port),
			connectionjsonrpc.WithLogin(cfg.GetZabbix().User),
			connectionjsonrpc.WithPasswd(cfg.GetAuthenticationData().ZabbixPasswd),
			connectionjsonrpc.WithConnectionTimeout(cmp.Or(cfg.GetZabbix().Timeout, 10)),
		)
	}
	if err != nil {
		log.Fatalf("error zabbix connection: %v", err)
	}
	// выполняем проверку доступности Zabbix путём попытки тестовой авторизации так как
	// для большей части методов требуется авторизация
	if err = zabbixConn.AuthorizationStart(ctx); err != nil {
		log.Fatalf("zabbix authorization error: %v", err)
	}

	//*********************************************************************************
	//********************** инициализация временного хранилища ***********************
	appStorage, err := appstorage.New(appstorage.WithSizeLogs(100))
	if err != nil {
		log.Fatalf("error initializing the temporary storage: %v", err)
	}
	// заполняем хранилище информацией о конфигурации, нужно для того что бы apiserver
	// мог отображать краткую информацию о настройках приложения
	appStorage.SetTaskSchedulerTimeJob(cfg.GetSchedule().TimerJob)
	dailyJobs := cfg.GetSchedule().DailyJob
	taskSchedulerTimeJobs := make([]string, 0, len(dailyJobs))
	for _, v := range dailyJobs {
		taskSchedulerTimeJobs = append(taskSchedulerTimeJobs, v)
	}
	appStorage.SetTaskSchedulerDailyJobs(taskSchedulerTimeJobs)
	appStorage.SetNetbox(appstorage.ShortParameters{
		Host: cfg.GetNetBox().Host,
		Port: cfg.GetNetBox().Port,
	})
	appStorage.SetZabbix(appstorage.ShortParameters{
		User: cfg.GetZabbix().User,
		Host: cfg.GetZabbix().Host,
		Port: cfg.GetZabbix().Port,
	})
	appStorage.SetDatabaseLogging(appstorage.ShortParameters{
		User: cfg.GetLogDB().User,
		Host: cfg.GetLogDB().Host,
		Port: cfg.GetLogDB().Port,
	})

	//*********************************************************************************
	//***************** инициализация обработчика логирования данных ******************
	logging := logginghandler.New(simpleLogger)
	logging.Start(ctx)

	//*********************************************************************************
	//******************** инициализация API сервера (web-server) *********************
	apiServer, err := apiserver.New(
		logging,
		appStorage,
		apiserver.WithHost(cfg.GetInformationServerApi().Host),
		apiserver.WithPort(cfg.GetInformationServerApi().Port),
		apiserver.WithAuthToken(cfg.GetAuthenticationData().APIServerToken),
		apiserver.WithVersion(versionApp),
	)
	if err != nil {
		log.Fatalf("error initializing the api server: %v", err)
	}
	// запуск сервера
	go apiServer.Start(ctx)

	// добавляем логирование в API сервер (вывод логов на веб-странице)
	logging.AddTransmittersFunc(func(msg interfaces.Messager) {
		appStorage.AddLog(appstorage.LogInformation{
			Date:        time.Now().Format(time.RFC3339),
			Type:        strings.ToUpper(msg.GetType()),
			Description: msg.GetMessage(),
		})
		if b, err := json.Marshal(struct {
			Type string `json:"type"`
			Data any    `json:"data"`
		}{
			Type: "logs",
			Data: appStorage.GetLogs(),
		}); err == nil {
			apiServer.SendData(b)
		}
	})

	//********************************************************************************
	//************************ инициализация клиента Netbox **************************
	nbCfg := cfg.GetNetBox()
	nbClient, err := netboxapi.New(
		cfg.AuthenticationData.NetBoxToken,
		netboxapi.WithHost(nbCfg.Host),
		netboxapi.WithPort(nbCfg.Port),
		netboxapi.WithTimeout(nbCfg.Timeout),
	)
	if err != nil {
		log.Fatalf("error initializing the netbox client: %v", err)
	}

	if err != nil {
		log.Fatalf("error initializing the Netbox client: %v", err)
	}

	//********************************************************************************
	//********************** инициализация обработчика задач *************************
	// это фактически то что будет выполнятся по рассписанию или при ручной инициализации
	taskHandlerSettings := taskhandlers.NewSettings(zabbixConn, nbClient, apiServer, appStorage, logging)
	taskHandler := taskHandlerSettings.Init(ctx)

	//********************************************************************************
	//******************** инициализация обработчика расписаний **********************
	sw, err := schedulehandler.NewScheduleHandler(
		schedulehandler.WithTimerJob(cfg.Schedule.TimerJob),
		schedulehandler.WithDailyJob(cfg.Schedule.DailyJob),
	)
	if err != nil {
		log.Fatalf("error module 'schedulehandler': '%v'", err)
	}
	// запуск автоматического обработчика заданий
	if err = sw.Start(
		ctx,
		func() {
			if err := taskHandler.SimpleTaskHandler(); err != nil {
				logging.Send("error", wrappers.WrapperError(err).Error())
			}
		}); err != nil {
		log.Fatalf("error start module 'schedulehandler': %v", err)
	}

	// запуск ручного обработчика заданий
	// отслеживает инициализацию выполнения задачи через веб-интерфейс
	taskHandler.TaskHandlerInitiatedThroughChannel()

	// получаем дополнительную информацию о Zabbix нужную для вывода в информационном сообщении
	b, err := zabbixConn.GetAPIInfo(ctx)
	if err != nil {
		log.Fatalf("error zabbix connection: %v", err)
	}
	zabbixApiInfo, errMsg, err := connectionjsonrpc.NewResponseAPIInfo().Get(b)
	if err != nil {
		simpleLogger.Write("error", wrappers.WrapperError(err).Error())
	}
	if errMsg != nil && errMsg.Error.Message != "" {
		simpleLogger.Write("warning", wrappers.WrapperError(err).Error())
	}

	//вывод информационного сообщения
	msg := getInformationMessage(cfg, zabbixApiInfo.Result)
	_ = simpleLogger.Write("info", msg)

	<-ctx.Done()

	sw.StopAllJobs()
	sw.Stop()
}
