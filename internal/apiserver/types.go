package apiserver

import (
	"net/http"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/sseserver"
)

// InformationServer информационный сервер
type InformationServer struct {
	logger                  interfaces.Logger
	storage                 interfaces.StorageInformation
	transmitterFromFrontend interfaces.BytesTransmitter
	transmitterToFrontend   interfaces.BytesTransmitter
	server                  *http.Server
	sseServer               *sseserver.SSEServer
	timeStart               time.Time
	timeout                 time.Duration
	host                    string
	version                 string
	port                    int
}

type informationServerOptions func(*InformationServer) error
