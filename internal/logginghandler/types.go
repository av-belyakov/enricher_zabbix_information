package logginghandler

import "github.com/av-belyakov/enricher_zabbix_information/interfaces"

type LoggingChan struct {
	dataWriter           interfaces.WriterLoggingData
	chanSystemMonitoring chan<- interfaces.Messager
	chanLogging          chan interfaces.Messager
}

// MessageLogging содержит информацию используемую при логировании
type MessageLogging struct {
	Message, Type string
}
