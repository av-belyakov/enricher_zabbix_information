package main

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/av-belyakov/simplelogger"
	zconnection "github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/dictionarieshandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/elasticsearchapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/logginghandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/schedulehandler"
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
	// ********************** инициализация модуля чтения словарей ********************
	dicts, err := dictionarieshandler.Read("config/dictionary.yml")
	if err != nil {
		log.Fatalf("error module 'dictionarieshandler': %v", err)
	}

	fmt.Println("Dictionaries:", dicts)

	//*********************************************************************************
	//***************** инициализация модуля-обёртки соединения с Zabbix **************
	zabbixConn, err := zconnection.NewConnect(
		zconnection.WithTLS(),
		zconnection.WithInsecureSkipVerify(),
		zconnection.WithHost(conf.GetZabbix().Host),
		zconnection.WithPort(conf.GetZabbix().Port),
		zconnection.WithLogin(conf.GetZabbix().User),
		zconnection.WithPasswd(conf.GetAuthenticationData().ZabbixPasswd),
		zconnection.WithConnectionTimeout(cmp.Or(conf.GetZabbix().Timeout, 10)),
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
	//***************** инициализация обработчика логирования данных ******************
	logging := logginghandler.New(simpleLogger)
	logging.Start(ctx)

	//******************************************************************
	//************** инициализация маршрутизатора данных ***************
	sw, err := schedulehandler.NewScheduleHandler(
		schedulehandler.WithTimerJob(conf.Schedule.TimerJob),
		schedulehandler.WithDailyJob(conf.Schedule.DailyJob),
	)
	if err != nil {
		log.Fatalf("error module 'schedulehandler': '%v'", err)
	}
	if err = sw.Start(
		ctx,
		func() error {
			/*

			   тут надо добавить обработчик который запускается по расписанию

			*/
			fmt.Println("START worker to 'schedule', current time:", time.Now())

			return nil
		}); err != nil {
		log.Fatalf("error start module 'schedulehandler': %v", err)
	}

	// получаем дополнительную информацию о Zabbix
	b, err := zabbixConn.GetAPIInfo(ctx)
	if err != nil {
		log.Fatalf("error zabbix connection: %v", err)
	}
	zabbixApiInfo, errMsg, err := zconnection.NewResponseAPIInfo().Get(b)
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
