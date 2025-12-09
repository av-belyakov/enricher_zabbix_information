package apiserver

import (
	"errors"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

//******************** функциональные настройки ***********************

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

// WithChToModule устанавливает канал для взаимодействия с модулем
func WithChToModule(v <-chan interfaces.BytesTransmitter) informationServerOptions {
	return func(whs *InformationServer) error {
		if v != nil {
			whs.chToFrontend = v

			return nil
		} else {
			return errors.New("the channel for interaction with the module must be initialized")
		}
	}
}

// WithChFromModule устанавливает канал для получения данных из модуля
func WithChFromModule(v chan<- interfaces.BytesTransmitter) informationServerOptions {
	return func(whs *InformationServer) error {
		if v != nil {
			whs.chFromFrontend = v

			return nil
		} else {
			return errors.New("the channel for receiving data from the module must be initialized")
		}
	}
}
