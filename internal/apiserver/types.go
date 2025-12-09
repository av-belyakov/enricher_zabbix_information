package apiserver

import (
	"net/http"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// InformationServer информационный сервер
type InformationServer struct {
	logger         interfaces.Logger
	storage        interfaces.StorageInformation
	server         *http.Server
	timeStart      time.Time
	timeout        time.Duration
	host           string
	version        string
	port           int
	chFromFrontend chan<- interfaces.BytesTransmitter
	chToFrontend   <-chan interfaces.BytesTransmitter
}

type informationServerOptions func(*InformationServer) error
