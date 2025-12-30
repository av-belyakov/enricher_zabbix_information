package taskhandlers

import (
	"context"

	zconnection "github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
)

// TaskHandler обработчик задач
type TaskHandler struct {
	settings   TaskHandlerSettings
	ctx        context.Context
	chanSignal chan<- ChanSignalSettings
}

type ChanSignalSettings struct {
	Data    []byte
	ForWhom string
}

// TaskHandlerSettings настройки обработчика задач
type TaskHandlerSettings struct {
	zabbixConn *zconnection.ZabbixConnectionJsonRPC
	apiServer  *apiserver.InformationServer
	storage    *storage.ShortTermStorage
	//netbox
	logger interfaces.Logger
}

// ResponseTaskHandler ответ обработчика задач
type ResponseTaskHandler struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// channelContextKey ключ для контекста
type channelContextKey string
