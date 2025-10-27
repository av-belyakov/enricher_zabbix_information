package logginghandler

import (
	"context"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// New конструктор обработки логов
// Это просто мост соединяющий несколько пакетов логирования
func New(writer interfaces.WriterLoggingData /*chSysMonit chan<- interfaces.Messager*/) *LoggingChan {
	return &LoggingChan{
		dataWriter: writer,
		//chanSystemMonitoring: chSysMonit,
		chanLogging: make(chan interfaces.Messager),
	}
}

// Start обработчик и распределитель логов
func (lc *LoggingChan) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case msg := <-lc.chanLogging:
				//передача информации по логам в simpleLogger (может так же писать логи в БД)
				//здесь так же может быть вывод в консоль, с индикацией цветов соответствующих
				//определенному типу сообщений но для этого надо включить вывод на stdout
				//в конфигурационном файле
				_ = lc.dataWriter.Write(msg.GetType(), msg.GetMessage())

				//здесь и далее возможна так же передача логов в некоторую систему
				// мониторинга, например Zabbix или дополнительно в какую либо БД
				//if msg.GetType() == "error" || msg.GetType() == "warning" {
				//	msg := NewMessageLogging()
				//	msg.SetType("error")
				//	msg.SetMessage(fmt.Sprintf("%s: %s", msg.GetType(), msg.GetMessage()))
				//
				//	lc.chanSystemMonitoring <- msg
				//}
				//if msg.GetType() == "info" {
				//	msg := NewMessageLogging()
				//	msg.SetType("info")
				//	msg.SetMessage(msg.GetMessage())
				//
				//	lc.chanSystemMonitoring <- msg
				//}
			}
		}
	}()
}
