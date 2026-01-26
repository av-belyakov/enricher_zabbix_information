package logginghandler

import (
	"context"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

// New конструктор обработки логов
// Это просто мост соединяющий несколько пакетов логирования
func New(writer interfaces.WriterLoggingData) *LoggingChan {
	return &LoggingChan{
		dataWriter:  writer,
		chanLogging: make(chan interfaces.Messager, 10),
	}
}

// AddTransmittersFunc дополнительные передатчики в которые будут передаваться логи
func (lc *LoggingChan) AddTransmittersFunc(functions ...func(msg interfaces.Messager)) {
	for _, f := range functions {
		lc.transmittersFunc = append(lc.transmittersFunc, f)
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

				for _, f := range lc.transmittersFunc {
					f(msg)
				}
			}
		}
	}()
}
