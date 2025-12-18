package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/dictionarieshandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

/*
		!!!!!!!
	TaskHandler не очень мне нравится когда все инициализированные модули
	свалены в одном месте. Думаю надо постараться как то обезличить, необходимые
	для выполнения задачи модули, с помощью интерфейсов и каналов
		!!!!!!
*/

func NewTaskHandler(
	zabbixConn *connectionjsonrpc.ZabbixConnectionJsonRPC,
	api *apiserver.InformationServer,
	storage *storage.ShortTermStorage,
	logger interfaces.Logger,
) *TaskHandlerSettings {
	return &TaskHandlerSettings{
		storage:    storage,
		zabbixConn: zabbixConn,
		apiServer:  api,
		logger:     logger,
		//netbox
	}
}

// autoTaskHandler автоматический обработчик задач, задачи запускаются по расписанию
func (ths *TaskHandlerSettings) AutoTaskHandler(ctx context.Context) error {
	return ths.start(ctx)
}

// manualTaskHandler ручной обработчик задач, задачи запускаются при их
// инициализации через веб-интерфейс
func (ths *TaskHandlerSettings) ManualTaskHandler(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case b := <-ths.apiServer.GetChannelOutgoingData():
				fmt.Println("TEST - Outgoing data from API server:", string(b))

				elmManualTask := apiserver.ElementManuallyTask{}
				if err := json.Unmarshal(b, &elmManualTask); err != nil {
					ths.logger.Send("error", wrappers.WrapperError(err).Error())

					continue
				}

				if elmManualTask.Type != "manually_task" {
					continue
				}

				// проверка токена доступа
				if !ths.apiServer.CheckAuthToken(elmManualTask.Settings.Token) {
					ths.logger.Send("error", wrappers.WrapperError(errors.New("invalid apiserver access token received")).Error())

					b, err := json.Marshal(apiserver.ElementManuallyTask{
						Type: "manually_task",
						Settings: apiserver.ElementManuallyTaskSettings{
							Error: "token invalide",
						},
					})
					if err != nil {
						ths.logger.Send("error", wrappers.WrapperError(err).Error())
					}

					// передаем результат на api сервер
					ths.apiServer.SendData(b)

					continue
				}

				if err := ths.start(ctx); err != nil {
					ths.logger.Send("error", wrappers.WrapperError(err).Error())
				}
			}
		}
	}()

	return nil
}

func (ths *TaskHandlerSettings) start(ctx context.Context) error {

	/*
		ну и тут все надо сделать, пока что тут все свалено в кучу

		порядок действий:
		1. считать словари;
		2. получить информацию из zabbix по хостам соответствующим пунктам словоря,
		если словари не найдены или пустые то получить информацию из zabbix по всем хостам;
		3. сохранить полученный из zabbix, результат во временном хранилище ShortTermStorage;
		4. запустить dnsresolver для преобразования доменных имен из информации
		о хостах, хранящейся в ShortTermStorage, в ip адреса, сохранить информацию
		во временном хранилище;
		5. выполнить поиск ip адресов в Netbox для того что бы проверить входят ли
		полученные ранее ip адреса в контролируемые сетевые диапазоны, при этом
		от Netbox нужно получить не только входят/не входят но и id сенсора который
		контролирует данный ip адрес;
		6. добавить теги содержащие id сенсора в информацию о хосте в Zabbix;
		7. если задача была инициализирована вручную, через веб-интерфейс, то
		отправить результат на api сервер.
	*/

	// инициализация словарей
	//
	// Думаю что чтение словарей должно быть каждый раз при запуске задачи, это
	// позволит измениять состав словарей не перезапуская приложение. Кроме того
	// следует обратить внимание на то что если словари не будут найдены или
	// они будут пустыми то из zabbix забираем все данные по хостам. Отсутствие
	// словарей не является критической ошибкой.
	// Таким образом место чтения словарей в обработчике задач.
	dicts, err := dictionarieshandler.Read("config/dictionary.yml")
	if err != nil {
		return err
	}

	//получаем полный список групп хостов
	res, err := ths.zabbixConn.GetFullHostGroupList(ctx)
	if err != nil {
		return err
	}

	data, errMsg, err := connectionjsonrpc.NewResponseGetHostGroupList().Get(res)
	if err != nil {
		return err
	}
	if errMsg.Error.Message != "" {
		ths.logger.Send("warning", wrappers.WrapperError(errors.New(errMsg.Error.Message)).Error())
	}

	//проверяем наличие списка групп хостов
	if len(data.Result) == 0 {
		return errors.New("an empty list of host groups has been received, no further processing of the task is possible")
	}

	if len(dicts.Dictionaries.WebSiteGroupMonitoring) != 0 {
		for _, host := range data.Result {
			if slices.ContainsFunc(dicts.Dictionaries.WebSiteGroupMonitoring, func(dict dictionarieshandler.Dictionaries) bool {
				if dict.Name == host.Name {
					return true
				}

				return false
			}) {

			}

			//получаем список хостов по группе
			res, err := ths.zabbixConn.GetHostListByGroup(ctx, dict.Name)
			if err != nil {
				return err
			}
		}
	}

	api.SendData(b)

	return nil
}
