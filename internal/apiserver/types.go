package apiserver

import (
	"net/http"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// InformationServer информационный сервер
type InformationServer struct {
	logger    interfaces.Logger
	server    *http.Server
	timeStart time.Time
	timeout   time.Duration
	host      string
	version   string
	port      int
	chTaskMng chan TaskManagementChannel
}

type informationServerOptions func(*InformationServer) error

type TaskManagementChannel struct {
	Type    string
	Command string
}
