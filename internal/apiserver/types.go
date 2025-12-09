package apiserver

import (
	"net/http"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/websocketserver"
)

// InformationServer информационный сервер
type InformationServer struct {
	logger                  interfaces.Logger
	storage                 interfaces.StorageInformation
	transmitterFromFrontend interfaces.BytesTransmitter
	transmitterToFrontend   interfaces.BytesTransmitter
	wsServer                *websocketserver.Hub
	server                  *http.Server
	timeStart               time.Time
	timeout                 time.Duration
	host                    string
	version                 string
	port                    int
}

type informationServerOptions func(*InformationServer) error
