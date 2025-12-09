package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/schedulehandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	zconnection "github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"
	"github.com/stretchr/testify/assert"
)

type AppCore struct {
	zabbixConnection *zconnection.ZabbixConnectionJsonRPC
	logging          interfaces.Logger
	config           *confighandler.ConfigApp
}

func NewAppCore(conf *confighandler.ConfigApp, zc *zconnection.ZabbixConnectionJsonRPC) {
	/*
	   продумать все же где будет инициализироваться api сервер, хранилище и
	   обработчик рассписаний
	*/
}

func (core *AppCore) Start(ctx context.Context) error {

	storageTemp := storage.NewShortTermStorage()
	api, err := apiserver.New(
		logging,
		storageTemp,
		apiserver.WithHost(host),
		apiserver.WithPort(port),
		apiserver.WithVersion(version),
	)
	assert.NoError(t, err)

	api.Start(ctx)

	//*********************************************************************************
	//******************** инициализация обработчика рассписаний **********************
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

							   	// !!!!! инициализация словарей !!!!!!
					// Думаю что чтение словарей должно быть каждый раз при запуске задачи, это
					// позволит измениять состав словарей не перезапуская приложение. Кроме того
					// следует обратить внимание на то что если словари не будут найдены или
					// они будут пустыми то из zabbix забираем все данные по хостам. Отсутствие
					// словарей не является критической ошибкой.
					// Таким образом место чтения словарей в обработчике задач.
					//dicts, err := dictionarieshandler.Read("config/dictionary.yml")
					//if err != nil {
					//	simpleLogger.Write("error", wrappers.WrapperError(err).Error())
					//}
					//fmt.Println("Dictionaries:", dicts)


			*/
			fmt.Println("START worker to 'schedule', current time:", time.Now())

			return nil
		}); err != nil {
		log.Fatalf("error start module 'schedulehandler': %v", err)
	}

}
