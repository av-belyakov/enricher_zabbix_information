package logginghandler

import (
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

type LoggingChan struct {
	transmittersFunc []func(msg interfaces.Messager)
	transmitters     []interfaces.BytesTransmitter
	dataWriter       interfaces.WriterLoggingData
	chanLogging      chan interfaces.Messager
}

// MessageLogging содержит информацию используемую при логировании
type MessageLogging struct {
	Message, Type string
}
