package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/av-belyakov/simplelogger"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/dictionarieshandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/elasticsearchapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/logginghandler"
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

	// ****************************************************************
	// ************* инициализация модуля чтения словарей *************
	dicts, err := dictionarieshandler.Read("config/dictionary.yml")
	if err != nil {
		log.Fatalf("error module 'dictionarieshandler': %v", err)
	}

	fmt.Println("Dictionaries:", dicts)

	// ***************************************************************************
	// ************* инициализация модуля взаимодействия с Service 1 *************

	// ***************************************************************************
	// ************* инициализация модуля взаимодействия с Service 2 *************

	//***************************************************************************
	//************** инициализация обработчика логирования данных ***************
	logging := logginghandler.New(simpleLogger)
	logging.Start(ctx)

	//******************************************************************
	//************** инициализация маршрутизатора данных ***************

	//вывод информационного сообщения
	msg := getInformationMessage(conf)
	_ = simpleLogger.Write("info", msg)

	<-ctx.Done()
}
