package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/customerrors"
	"github.com/av-belyakov/enricher_zabbix_information/internal/dnsresolver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/netboxapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

func NewSettings(
	zabbixConn *connectionjsonrpc.ZabbixConnectionJsonRPC,
	netboxClient *netboxapi.Client,
	apiServer *apiserver.InformationServer,
	storage *storage.ShortTermStorage,
	logger interfaces.Logger,
) *TaskHandlerSettings {
	return &TaskHandlerSettings{
		netboxClient: netboxClient,
		zabbixConn:   zabbixConn,
		apiServer:    apiServer,
		storage:      storage,
		logger:       logger,
	}
}

// Init инициализация обработчика задач
func (ths TaskHandlerSettings) Init(ctx context.Context) *TaskHandler {
	chanSignal := make(chan ChanSignalSettings)

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(chanSignal)

				return

			case msg := <-chanSignal:
				if msg.ForWhom == "web" {
					ths.apiServer.SendData(msg.Data)
				}
			}
		}
	}()

	return &TaskHandler{
		settings:   ths,
		chanSignal: chanSignal,
		ctx:        ctx,
	}
}

// SimpleTaskHandler простой обработчик задач
func (th *TaskHandler) SimpleTaskHandler() error {
	if err := th.start(); err != nil {
		return wrappers.WrapperError(err)
	}

	return nil
}

// TaskHandlerInitiatedThroughChannel обработчик задач, задачи запускаются
// при их инициализации через веб-интерфейс
func (th *TaskHandler) TaskHandlerInitiatedThroughChannel() error {
	go func() {
		for msg := range th.settings.apiServer.GetChannelOutgoingData() {
			elmManualTask := apiserver.ElementManuallyTask{}
			if err := json.Unmarshal(msg, &elmManualTask); err != nil {
				th.settings.logger.Send("error", wrappers.WrapperError(err).Error())

				continue
			}

			if elmManualTask.Type != "manually_task" {
				continue
			}

			// проверка токена доступа
			if !th.settings.apiServer.CheckAuthToken(elmManualTask.Settings.Token) {
				th.settings.logger.Send("error", wrappers.WrapperError(errors.New("invalid apiserver access token received")).Error())

				b, err := json.Marshal(apiserver.ElementManuallyTask{
					Type: "manually_task",
					Settings: apiserver.ElementManuallyTaskSettings{
						Error: "token invalide",
					},
				})
				if err != nil {
					th.settings.logger.Send("error", wrappers.WrapperError(err).Error())

					continue
				}

				th.chanSignal <- ChanSignalSettings{
					ForWhom: "web",
					Data:    b,
				}

				continue
			}

			if err := th.start(); err != nil {
				th.settings.logger.Send("error", wrappers.WrapperError(err).Error())
			}
		}
	}()

	return nil
}

func (th *TaskHandler) start() error {
	// получаем полный список групп хостов Zabbix
	res, err := th.settings.zabbixConn.GetFullHostGroupList(th.ctx)
	if err != nil {
		return err
	}

	hostGroupList, errMsg, err := connectionjsonrpc.NewResponseGetHostGroupList().Get(res)
	if err != nil {
		return err
	}
	if errMsg.Error.Message != "" {
		th.settings.logger.Send("warning", errMsg.Error.Message)
	}

	// проверяем наличие списка групп хостов
	if len(hostGroupList.Result) == 0 {
		return errors.New("an empty list of host groups has been received, no further processing of the task is possible")
	}

	// получаем список id групп хостов, которые нужно отслеживать, на основе корреляции
	// имен веб-сайтов предназначенных для мониторинга
	//
	// Думаю, что чтение словарей должно быть каждый раз при запуске задачи, это
	// позволит изменять состав словарей не перезапуская приложение. Кроме того,
	// следует обратить внимание на то, что если словари не будут найдены или
	// они будут пустыми, то из zabbix забираем все данные по хостам. Отсутствие
	// словарей не является критической ошибкой.
	listGroupsId, err := GetListIdsWebsitesGroupMonitoring(constants.App_Dictionary_Path, hostGroupList.Result)
	if err != nil {
		th.settings.logger.Send("error", wrappers.WrapperError(err).Error())
	}

	//fmt.Println("method 'TaskHandlerSettings.start' listGroupsId:", listGroupsId)

	// получаем список хостов которые есть в словарях, если словари
	// не пусты, или все хосты
	res, err = th.settings.zabbixConn.GetHostList(th.ctx, listGroupsId...)
	if err != nil {
		return err
	}

	hostList, errMsg, err := connectionjsonrpc.NewResponseGetHostList().Get(res)
	if err != nil {
		return err
	}

	//fmt.Println("method 'TaskHandlerSettings.start' hostList:", hostList)

	//очищаем хранилище от предыдущих данных (что бы не смешивать старые и новые данные)
	th.settings.storage.DeleteAll()
	// устанавливаем дату начала выполнения задачи
	th.settings.storage.SetStartDateExecution()

	// заполняем хранилище данными о хостах
	for _, host := range hostList.Result {
		if hostId, err := strconv.Atoi(host.HostId); err == nil {
			th.settings.storage.Add(storage.HostDetailedInformation{
				HostId:       hostId,
				OriginalHost: host.Host,
			})
		}
	}

	//fmt.Printf("method 'TaskHandlerSettings.start' count hosts:'%d'\n", len(ths.storage.GetList()))

	// инициализируем поиск ip адресов через DNS resolver
	dnsRes, err := dnsresolver.New(
		th.settings.storage,
		dnsresolver.WithTimeout(10),
	)
	if err != nil {
		return err
	}

	// меняем статус задачи на "выполняется"
	th.settings.storage.SetProcessRunning()

	// запускаем поиск через DNS resolver
	chInfo, err := dnsRes.Run(th.ctx)
	if err != nil {
		return err
	}

	for msg := range chInfo {
		if err := th.settings.storage.SetIsProcessed(msg.HostId); err != nil {
			th.settings.logger.Send("error", wrappers.WrapperError(err).Error())
		}

		if msg.Error != nil {
			if err := th.settings.storage.SetError(msg.HostId, customerrors.NewErrorNoValidUrl(msg.OriginalHost, err)); err != nil {
				th.settings.logger.Send("error", wrappers.WrapperError(err).Error())
			}

			continue
		}

		if err := th.settings.storage.SetDomainName(msg.HostId, msg.DomainName); err != nil {
			th.settings.logger.Send("error", wrappers.WrapperError(err).Error())
		}

		if err := th.settings.storage.SetIps(msg.HostId, msg.Ips[0], msg.Ips...); err != nil {
			th.settings.logger.Send("error", wrappers.WrapperError(err).Error())
		}

		b, err := json.Marshal(ResponseTaskHandler{
			Type: "ask_manually_task",
			Data: supportingfunctions.CreateTaskStatistics(th.settings.storage),
		})
		if err != nil {
			th.settings.logger.Send("error", wrappers.WrapperError(err).Error())

			continue
		}

		th.chanSignal <- ChanSignalSettings{
			ForWhom: "web",
			Data:    b,
		}
	}

	// логируем ошибки при выполнении DNS преобразования доменных имён в ip адреса
	errList := th.settings.storage.GetListErrors()
	for _, v := range errList {
		th.settings.logger.Send("warning", fmt.Sprintf("error DNS resolve '%s', description:'%s'", v.OriginalHost, v.Error.Error()))
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
	th.settings.storage.SetProcessNotRunning()

	b, err := json.Marshal(ResponseTaskHandler{
		Type: "ask_manually_task",
		Data: supportingfunctions.CreateTaskStatistics(th.settings.storage),
	})
	if err != nil {
		return err
	}

	th.chanSignal <- ChanSignalSettings{
		ForWhom: "web",
		Data:    b,
	}

	return nil
}
