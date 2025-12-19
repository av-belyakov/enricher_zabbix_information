package main

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/av-belyakov/simplelogger"
	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appversion"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/elasticsearchapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/logginghandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/schedulehandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

func app(ctx context.Context) {
	var nameRegionalObject string
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
	conf, err := confighandler.New(rootPath)
	if err != nil {
		log.Fatalf("error module 'confighandler': %v", err)
	}

	// ****************************************************************************
	// ********************* инициализация модуля логирования *********************
	var listLog []simplelogger.OptionsManager
	for _, v := range conf.GetListLogs() {
		listLog = append(listLog, v)
	}
	opts := simplelogger.CreateOptions(listLog...)
	simpleLogger, err := simplelogger.NewSimpleLogger(ctx, constants.Root_Dir, opts)
	if err != nil {
		log.Fatalf("error module 'simplelogger': %v", err)
	}

	//*********************************************************************************
	//********** инициализация модуля взаимодействия с БД для передачи логов **********
	confDB := conf.GetLogDB()
	if esc, err := elasticsearchapi.NewElasticsearchConnect(elasticsearchapi.Settings{
		Port:               confDB.Port,
		Host:               confDB.Host,
		User:               confDB.User,
		Passwd:             conf.GetAuthenticationData().WriteLogBDPasswd,
		IndexDB:            confDB.StorageNameDB,
		NameRegionalObject: nameRegionalObject,
	}); err != nil {
		_ = simpleLogger.Write("error", wrappers.WrapperError(err).Error())
	} else {
		//подключение логирования в БД
		simpleLogger.SetDataBaseInteraction(esc)
	}

	//*********************************************************************************
	//***************** инициализация модуля-обёртки соединения с Zabbix **************
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
		log.Fatalf("error zabbix connection: %v", err)
	}
	// выполняем проверку доступности Zabbix путём попытки тестовой авторизации так как
	// для большей части методов требуется авторизация
	if err = zabbixConn.AuthorizationStart(ctx); err != nil {
		log.Fatalf("authorization error: %v", err)
	}

	//*********************************************************************************
	//********************** инициализация временного хранилища ***********************
	storageTemp := storage.NewShortTermStorage()

	//*********************************************************************************
	//***************** инициализация обработчика логирования данных ******************
	logging := logginghandler.New(simpleLogger)
	logging.Start(ctx)

	//*********************************************************************************
	//******************** инициализация API сервера (web-server) *********************
	api, err := apiserver.New(
		logging,
		storageTemp,
		apiserver.WithHost(conf.GetInformationServerApi().Host),
		apiserver.WithPort(conf.GetInformationServerApi().Port),
		apiserver.WithAuthToken(conf.GetAuthenticationData().APIServerToken),
		apiserver.WithVersion(versionApp),
	)
	if err != nil {
		log.Fatalf("error initializing the api server: %v", err)
	}
	// запуск сервера
	go api.Start(ctx)

	// добавляем логирование в API сервер (вывод логов на веб-странице)
	logging.AddTransmitters(api)

	//********************************************************************************
	//********************** инициализация обработчика задач *************************
	// это фактически то что будет выполнятся по рассписанию или при ручной инициализации
	taskHandler := NewTaskHandler(zabbixConn, api, storageTemp, logging)

	//********************************************************************************
	//******************** инициализация обработчика расписаний **********************
	sw, err := schedulehandler.NewScheduleHandler(
		schedulehandler.WithTimerJob(conf.Schedule.TimerJob),
		schedulehandler.WithDailyJob(conf.Schedule.DailyJob),
	)
	if err != nil {
		log.Fatalf("error module 'schedulehandler': '%v'", err)
	}
	// запуск автоматического обработчика заданий
	if err = sw.Start(
		ctx,
		func() {
			if err := taskHandler.AutoTaskHandler(ctx); err != nil {
				logging.Send("error", wrappers.WrapperError(err).Error())
			}
		}); err != nil {
		log.Fatalf("error start module 'schedulehandler': %v", err)
	}

	// запуск ручного обработчика заданий
	// отслеживает инициализацию выполнения задачи через веб-интерфейс
	taskHandler.ManualTaskHandler(ctx)

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
	msg := getInformationMessage(conf, zabbixApiInfo.Result)
	_ = simpleLogger.Write("info", msg)

	<-ctx.Done()

	sw.StopAllJobs()
	sw.Stop()
}
