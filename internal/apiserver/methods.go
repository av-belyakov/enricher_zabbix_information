package apiserver

import "time"

// SendCommandStartProcess отправляет команду на запуск процесса
func (is *InformationServer) SendCommandStartProcess() {
	is.chTaskMng <- TaskManagementChannel{
		Type:    "task processing",
		Command: "start process",
	}
}

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
