package helpers

import "github.com/av-belyakov/enricher_zabbix_information/interfaces"

// LoggingForTest тестовая структура для логирования
type LoggingForTest struct {
	chMessage chan interfaces.Messager
}

// MessageForTest тестовая структура для описания сообщения
type MessageForTest struct {
	msgType, msgData string
}
