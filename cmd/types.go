package main

import (
	zconnection "github.com/av-belyakov/zabbixapicommunicator/v2/cmd/connectionjsonrpc"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
)

type TaskHandlerSettings struct {
	zabbixConn *zconnection.ZabbixConnectionJsonRPC
	apiServer  *apiserver.InformationServer
	storage    *storage.ShortTermStorage
	//netbox
	logger interfaces.Logger
}
