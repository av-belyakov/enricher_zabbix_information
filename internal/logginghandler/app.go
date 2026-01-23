package logginghandler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/shortlogstory"
)

// New конструктор обработки логов
// Это просто мост соединяющий несколько пакетов логирования
func New(writer interfaces.WriterLoggingData, storageLog *shortlogstory.ShortLogStory) *LoggingChan {
	return &LoggingChan{
		storage:     storageLog,
		dataWriter:  writer,
		chanLogging: make(chan interfaces.Messager, 10),
	}
}

// AddTransmitters дополнительные передатчики в которые будут передаваться логи
func (lc *LoggingChan) AddTransmitters(transmitters ...interfaces.BytesTransmitter) {
	for _, transmitting := range transmitters {
		lc.transmitters = append(lc.transmitters, transmitting)
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

				lc.storage.Add(shortlogstory.LogInformation{
					Date:        time.Now().Format(time.RFC3339),
					Type:        strings.ToUpper(msg.GetType()),
					Description: msg.GetMessage(),
				})

				fmt.Println("___ Logging message: ", msg.GetMessage(), " type: ", msg.GetType())

				for _, transmiting := range lc.transmitters {
					if b, err := json.Marshal(struct {
						Type string `json:"type"`
						Data any    `json:"data"`
					}{
						Type: "logs",
						Data: lc.storage.Get(),
					}); err == nil {
						transmiting.SendData(b)
					}
				}
			}
		}
	}()
}
