package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appstorage"
	"github.com/av-belyakov/enricher_zabbix_information/internal/customerrors"
	"github.com/av-belyakov/enricher_zabbix_information/internal/dnsresolver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/netboxapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

func NewSettings(
	zabbixConn *connectionjsonrpc.ZabbixConnectionJsonRPC,
	netboxClient *netboxapi.Client,
	apiServer *apiserver.InformationServer,
	storage *appstorage.SharedAppStorage,
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
				chanSignal = nil

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

				if th.chanSignal == nil {
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
	//очищаем хранилище от предыдущих данных (что бы не смешивать старые и новые данные)
	th.settings.storage.DeleteAll()

	// устанавливаем дату начала выполнения задачи
	th.settings.storage.SetStartDateExecution()
	// меняем статус задачи на "выполняется"
	th.settings.storage.SetProcessRunning()
	// меняем статус задачи на "не выполняется"
	defer th.settings.storage.SetProcessNotRunning()

	// получаем полный список групп хостов Zabbix
	res, err := th.settings.zabbixConn.GetFullHostGroupList(th.ctx)
	if err != nil {
		return err
	}

	// преобразуем список групп хостов Zabbix из бинарного вида в соответствующую структуру
	hostGroupList, errMsg, err := connectionjsonrpc.NewResponseGetHostGroupList().Get(res)
	if err != nil {
		return err
	}
	if errMsg.Error.Message != "" {
		th.settings.logger.Send("warning", errMsg.Error.Message)
	}

	// количество групп хостов в Zabbix
	th.settings.storage.SetCountZabbixHostsGroup(len(hostGroupList.Result))

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

	// количество групп хостов относящихся к веб-сайтам мониторинга
	th.settings.storage.SetCountMonitoringHostsGroup(len(listGroupsId))

	// получаем список хостов которые есть в словарях (имена веб-сайтов предназначенных
	// для мониторинга), если словари не пусты, или все хосты
	res, err = th.settings.zabbixConn.GetHostList(th.ctx, listGroupsId...)
	if err != nil {
		return err
	}

	// преобразуем список хостов Zabbix из бинарного вида в JSON
	hostList, errMsg, err := connectionjsonrpc.NewResponseGetHostList().Get(res)
	if err != nil {
		return err
	}

	// количество хостов по которым осуществляется мониторинг
	th.settings.storage.SetCountMonitoringHosts(len(hostList.Result))

	// заполняем хранилище данными о хостах
	for _, host := range hostList.Result {
		if hostId, err := strconv.Atoi(host.HostId); err == nil {
			th.settings.storage.AddElement(appstorage.HostDetailedInformation{
				HostId:       hostId,
				OriginalHost: host.Host,
			})
		}
	}

	// инициализируем поиск ip адресов через DNS resolver
	dnsRes, err := dnsresolver.New(dnsresolver.WithTimeout(10))
	if err != nil {
		return err
	}

	// запускаем поиск через DNS resolver
	chInfo, err := dnsRes.Run(th.ctx, th.settings.storage.GetHosts())
	if err != nil {
		return err
	}

	for msg := range chInfo {
		if msg.Error != nil {
			if err := th.settings.storage.SetError(msg.HostId, customerrors.NewErrorNoValidUrl(msg.OriginalHost, err)); err != nil {
				th.settings.logger.Send("error", wrappers.WrapperError(err).Error())
			}

			continue
		}

		if err := th.settings.storage.SetDomainName(msg.HostId, msg.DomainName); err != nil {
			th.settings.logger.Send("error", wrappers.WrapperError(err).Error())
		}

		if err := th.settings.storage.SetIps(msg.HostId, msg.Ips...); err != nil {
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

		// передача информации на веб-интерфейс
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

	// получаем префиксы из Netbox
	chunPrefixInfo, countPrefixes, err := NetboxPrefixes(th.ctx, th.settings.netboxClient, th.settings.logger)
	if err != nil {
		return err
	}
	if countPrefixes == 0 {
		return errors.New("an empty list of prefixes (subnets) was received from the netbox")
	}

	// количество найденных префиксов в Netbox
	th.settings.storage.SetCountNetboxPrefixes(countPrefixes)

	/*
		нужен ещё один параметр мтатистики - количество полученных (или
		может быть обработанных) префиксов Netbox

		тогда можно будет отслеживать ход получения информаии о префиксах
		из Netbox, это самая затратная по времени операция
	*/

	shortPrefixList := netboxapi.ShortPrefixList{}
	for prefixInfo := range chunPrefixInfo {
		shortPrefixList = append(shortPrefixList, prefixInfo...)

		th.settings.storage.SetCountNetboxPrefixesReceived(int(th.settings.storage.GetCountNetboxPrefixesReceived()) + len(prefixInfo))
		if b, err := json.Marshal(ResponseTaskHandler{
			Type: "ask_manually_task",
			Data: supportingfunctions.CreateTaskStatistics(th.settings.storage),
		}); err == nil {
			th.chanSignal <- ChanSignalSettings{
				ForWhom: "web",
				Data:    b,
			}
		}
	}

	//shortPrefixList := GetNetboxPrefixes(th.ctx, th.settings.netboxClient, th.settings.logger)
	//if shortPrefixList.Count == 0 {
	//	return errors.New("an empty list of prefixes (subnets) was received from the netbox")
	//}

	// выполняем поиск ip адресов в префиксах полученных от Netbox
	// в SearchIpaddrToPrefixesNetbox передаётся хранилище в котором выполняются
	// все изменения произошедшие в результате поиска ip адресов в префиксах
	SearchIpaddrToPrefixesNetbox(runtime.NumCPU(), th.settings.storage, shortPrefixList, th.settings.logger)

	var num int
	// добавляем или обновляем теги в Zabbix
	for _, v := range th.settings.storage.GetHostsWithSensorId() {
		var sensorsId string
		if len(v.SensorsId) == 0 {
			continue
		} else if len(v.SensorsId) == 1 {
			sensorsId = v.SensorsId[0]
		} else {
			sensorsId = strings.Join(v.SensorsId, ",")
		}

		if _, err := th.settings.zabbixConn.UpdateHostParameterTags(
			th.ctx,
			fmt.Sprint(v.HostId),
			connectionjsonrpc.Tags{
				Tag: []connectionjsonrpc.Tag{
					{Tag: "СОА", Value: sensorsId},
				},
			},
		); err != nil {
			th.settings.logger.Send("error", wrappers.WrapperError(err).Error())

			continue
		}

		num++
	}

	// количество обновленных хостов в Zabbix
	th.settings.storage.SetCountUpdatedZabbixHosts(num)
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
