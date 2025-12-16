package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
	"github.com/av-belyakov/enricher_zabbix_information/internal/apiserver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/dictionarieshandler"
	"github.com/stretchr/testify/assert"
)

/*
		!!!!!!!
	TaskHandler не очень мне нравится когда все инициализированные модули
	свалены в одном месте. Думаю надо постараться как то обезличить, необходимые
	для выполнения задачи модули, с помощью интерфейсов и каналов
		!!!!!!
*/

func NewTaskHandler(logger interfaces.Logger, api *apiserver.InformationServer) *TaskHandlerSettings {
	return &TaskHandlerSettings{
		apiServer: api,
		logger:    logger,
	}
}

func (ths *TaskHandlerSettings) majorTask(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case b := <-ths.apiServer.GetChannelOutgoingData():

				/*
					ну и тут все надо сделать, пока что тут все свалено в кучу
				*/

				fmt.Println("TEST - Outgoing data from API server:", string(b))

				elmManualTask := apiserver.ElementManuallyTask{}

				err := json.Unmarshal(b, &elmManualTask)
				assert.NoError(t, err)

				if elmManualTask.Type != "manually_task" {
					continue
				}

				var strErr string
				// проверка токена доступа
				if !ths.apiServer.CheckAuthToken(elmManualTask.Settings.Token) {
					strErr = "token invalide"
				}

				b, err = json.Marshal(apiserver.ElementManuallyTask{
					Type: "manually_task",
					Settings: apiserver.ElementManuallyTaskSettings{
						Error: strErr,
					},
				})
				assert.NoError(t, err)

				// инициализация словарей
				//
				// Думаю что чтение словарей должно быть каждый раз при запуске задачи, это
				// позволит измениять состав словарей не перезапуская приложение. Кроме того
				// следует обратить внимание на то что если словари не будут найдены или
				// они будут пустыми то из zabbix забираем все данные по хостам. Отсутствие
				// словарей не является критической ошибкой.
				// Таким образом место чтения словарей в обработчике задач.
				dicts, err := dictionarieshandler.Read("config/dictionary.yml")
				if err != nil {
					//	simpleLogger.Write("error", wrappers.WrapperError(err).Error())
				}
				//fmt.Println("Dictionaries:", dicts)

				//
				//
				//   тут надо добавить обработчик который запускается по расписанию
				//
				//
				fmt.Println("START worker to 'schedule', current time:", time.Now())

				api.SendData(b)
			}
		}
	}()

	return nil
}
