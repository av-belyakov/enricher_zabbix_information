package apiserver

import (
	"net/http"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appstorage"
)

// InformationServer информационный сервер
type InformationServer struct {
	logger interfaces.Logger
	//storage   interfaces.StorageInformation
	storage   *appstorage.SharedAppStorage
	server    *http.Server
	timeStart time.Time
	timeout   time.Duration
	host      string
	version   string
	authToken string
	port      int
	chInput   chan []byte // канал для данных входящих в модуль
	chOutput  chan []byte // канал для данных изходящих из модуля
}

type informationServerOptions func(*InformationServer) error

// ElementManuallyTask ручной запуск задачи
type ElementManuallyTask struct {
	Type     string                      `json:"type"`
	Settings ElementManuallyTaskSettings `json:"settings"`
}

// ElementManuallyTaskSettings параметры ручного запуска задач
type ElementManuallyTaskSettings struct {
	Command string `json:"command"`
	Error   string `json:"error"`
	Token   string `json:"token"`
}
