package apiserver

import (
	"errors"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

func New(logger interfaces.Logger, storage interfaces.StorageInformation, opts ...informationServerOptions) (*InformationServer, error) {
	is := &InformationServer{
		version:   "0.0.1",
		timeStart: time.Now(),
		host:      "127.0.0.1",
		port:      7575,
		timeout:   time.Second * 10,
		logger:    logger,
		storage:   storage,
		chInput:   make(chan []byte),
		chOutput:  make(chan []byte),
	}
	for _, opt := range opts {
		if err := opt(is); err != nil {
			return is, err
		}
	}

	return is, nil
}

// WithTimeout устанавливает время ожидания выполнения запроса
func WithTimeout(v int) informationServerOptions {
	return func(is *InformationServer) error {
		is.timeout = time.Duration(v) * time.Second

		return nil
	}
}

// WithPort устанавливает порт для взаимодействия с модулем
func WithPort(v int) informationServerOptions {
	return func(is *InformationServer) error {
		is.port = v

		return nil
	}
}

// WithHost устанавливает хост для взаимодействия с модулем
func WithHost(v string) informationServerOptions {
	return func(is *InformationServer) error {
		is.host = v

		return nil
	}
}

// WithAuthTokn устанавливает авторизационный токен
// это позволяет взаимодействовать с модулем из веб-интерфейса
func WithAuthToken(v string) informationServerOptions {
	return func(is *InformationServer) error {
		if v == "" {
			return errors.New("the authorization token cannot be empty")
		}

		is.authToken = v

		return nil
	}
}

// WithVersion устанавливает версию модуля (опционально)
func WithVersion(v string) informationServerOptions {
	return func(is *InformationServer) error {
		is.version = v

		return nil
	}
}
