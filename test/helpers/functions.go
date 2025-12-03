package helpers

import "github.com/av-belyakov/enricher_zabbix_information/interfaces"

// NewLoggingForTest тестовый логгер
func NewLoggingForTest() *LoggingForTest {
	return &LoggingForTest{
		chMessage: make(chan interfaces.Messager),
	}
}
