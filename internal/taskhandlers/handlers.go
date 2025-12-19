package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/dictionarieshandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/dnsresolver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

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
func (ths *TaskHandlerSettings) TaskHandler(ctx context.Context) error {
	return wrappers.WrapperError(ths.start(ctx))
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
	//получаем полный список групп хостов
	res, err := ths.zabbixConn.GetFullHostGroupList(ctx)
	if err != nil {
		return err
	}

	hostGroupList, errMsg, err := connectionjsonrpc.NewResponseGetHostGroupList().Get(res)
	if err != nil {
		return err
	}
	if errMsg.Error.Message != "" {
		ths.logger.Send("warning", errMsg.Error.Message)
	}

	//fmt.Println("method 'TaskHandlerSettings.start' full host group list:")
	//for k, host := range hostGroupList.Result {
	//	fmt.Printf("%d.\n\tname:'%s'\n", k+1, host.Name)
	//}

	//проверяем наличие списка групп хостов
	if len(hostGroupList.Result) == 0 {
		return errors.New("an empty list of host groups has been received, no further processing of the task is possible")
	}

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
		ths.logger.Send("error", wrappers.WrapperError(err).Error())
	}
	dictsSize := len(dicts.Dictionaries.WebSiteGroupMonitoring)

	//fmt.Println("method 'TaskHandlerSettings.start' dictsSize:", dictsSize)

	var listGroupsId []string
	for _, host := range hostGroupList.Result {
		if dictsSize == 0 {
			listGroupsId = append(listGroupsId, host.GroupId)

			continue
		}

		if !slices.ContainsFunc(
			dicts.Dictionaries.WebSiteGroupMonitoring,
			func(v dictionarieshandler.WebSiteMonitoring) bool {
				if v.Name == host.Name {
					return true
				}

				return false
			},
		) {
			continue
		}

		listGroupsId = append(listGroupsId, host.GroupId)
	}

	fmt.Println("method 'TaskHandlerSettings.start' listGroupsId:", listGroupsId)

	// получаем список хостов или которые есть в словарях, если словари
	// не пусты, или все хосты
	res, err = ths.zabbixConn.GetHostList(ctx, listGroupsId...)
	if err != nil {
		return err
	}

	hostList, errMsg, err := connectionjsonrpc.NewResponseGetHostList().Get(res)
	if err != nil {
		return err
	}

	fmt.Println("method 'TaskHandlerSettings.start' hostList:", hostList)

	//очищаем хранилище от предыдущих данных (что бы не смешивать старые и новые данные)
	ths.storage.DeleteAll()
	// устанавливаем дату начала выполнения задачи
	ths.storage.SetStartDateExecution()

	// заполняем хранилище данными о хостах
	for _, host := range hostList.Result {
		if hostId, err := strconv.Atoi(host.HostId); err == nil {
			ths.storage.Add(datamodels.HostDetailedInformation{
				HostId:       hostId,
				OriginalHost: host.Host,
			})
		}
	}

	fmt.Printf("method 'TaskHandlerSettings.start' count hosts:'%d'\n", len(ths.storage.GetList()))

	// инициализируем поиск через DNS resolver
	dnsRes, err := dnsresolver.New(
		ths.storage,
		ths.logger,
		dnsresolver.WithTimeout(10),
	)
	if err != nil {
		return err
	}

	// меняем статус задачи на "выполняется"
	ths.storage.SetProcessRunning()

	chDone := make(chan struct{})
	// запускаем поиск через DNS resolver
	go dnsRes.Run(ctx, chDone)
	<-chDone

	// логируем ошибки при выполнении DNS преобразования доменных имён в ip адреса
	errList := ths.storage.GetListErrors()
	for _, v := range errList {
		ths.logger.Send("warning", fmt.Sprintf("error DNS resolve '%s', description:'%s'", v.OriginalHost, v.Error.Error()))
	}

	/*
		Далее нужно:
			1. выполнить поиск ip адресов в Netbox для того что бы проверить входят ли
			полученные ранее ip адреса в контролируемые сетевые диапазоны, при этом
			от Netbox нужно получить не только входят/не входят но и id сенсора который
			контролирует данный ip адрес;
			2. добавить теги содержащие id сенсора в информацию о хосте в Zabbix;
			3. если задача была инициализирована вручную, через веб-интерфейс, то
			отправить результат на api сервер.

	*/

	// меняем статус задачи на "не выполняется"
	ths.storage.SetProcessNotRunning()

	//api.SendData(b)

	return nil
}
