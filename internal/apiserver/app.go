package apiserver

import (
	"errors"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/sseserver"
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
		sseServer: sseserver.New(logger, storage),
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
	return func(whs *InformationServer) error {
		whs.timeout = time.Duration(v) * time.Second

		return nil
	}
}

// WithPort устанавливает порт для взаимодействия с модулем
func WithPort(v int) informationServerOptions {
	return func(whs *InformationServer) error {
		whs.port = v

		return nil
	}
}

// WithHost устанавливает хост для взаимодействия с модулем
func WithHost(v string) informationServerOptions {
	return func(whs *InformationServer) error {
		whs.host = v

		return nil
	}
}

// WithVersion устанавливает версию модуля (опционально)
func WithVersion(v string) informationServerOptions {
	return func(whs *InformationServer) error {
		whs.version = v

		return nil
	}
}

// WithTransmitterToModule устанавливает интерфейс для взаимодействия с модулем
func WithChToModule(v interfaces.BytesTransmitter) informationServerOptions {
	return func(whs *InformationServer) error {
		if v != nil {
			whs.transmitterToFrontend = v

			return nil
		} else {
			return errors.New("the transmitter for interaction with the module must be initialized")
		}
	}
}

// WithTransmitterFromModule устанавливает интерфейс для получения данных из модуля
func WithChFromModule(v interfaces.BytesTransmitter) informationServerOptions {
	return func(whs *InformationServer) error {
		if v != nil {
			whs.transmitterFromFrontend = v

			return nil
		} else {
			return errors.New("the transmitter for receiving data from the module must be initialized")
		}
	}
}
