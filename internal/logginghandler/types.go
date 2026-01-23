package logginghandler

import (
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/shortlogstory"
)

type LoggingChan struct {
	transmitters []interfaces.BytesTransmitter
	dataWriter   interfaces.WriterLoggingData
	storage      *shortlogstory.ShortLogStory
	chanLogging  chan interfaces.Messager
}

// MessageLogging содержит информацию используемую при логировании
type MessageLogging struct {
	Message, Type string
}
