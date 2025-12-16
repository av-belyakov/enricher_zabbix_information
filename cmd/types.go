package main

import (
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
)

type TaskHandlerSettings struct {
	apiServer  *apiserver.InformationServer
	logger     interfaces.Logger
	authTokent string
}
