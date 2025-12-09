package apiserver

import (
	"net/http"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// InformationServer информационный сервер
type InformationServer struct {
	logger  interfaces.Logger
	storage interfaces.StorageInformation
	//transmitterFromFrontend interfaces.BytesTransmitter
	//transmitterToFrontend   interfaces.BytesTransmitter
	server    *http.Server
	timeStart time.Time
	timeout   time.Duration
	host      string
	version   string
	port      int
	chInput   chan []byte // канал для данных входящих в модуль
	chOutput  chan []byte // канал для данных изходящих из модуля
}

type informationServerOptions func(*InformationServer) error
