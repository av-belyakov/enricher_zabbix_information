package schedule

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/subosito/gotenv"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/schedulehandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
)

func TestScheduleHandler(t *testing.T) {
	os.Setenv("GO_ENRICHERZI_MAIN", "test")

	if err := gotenv.Load(".env"); err != nil {
		t.Fatalf("Не удалось загрузить переменные окружения из файла .env: %v", err)
	}

	rootPath, err := supportingfunctions.GetRootPath(constants.Root_Dir)
	if err != nil {
		t.Fatalf("Не удалось получить корневую директорию: %v", err)
	}

	conf, err := confighandler.New(rootPath)
	if err != nil {
		t.Fatalf("Не удалось прочитать конфигурационный файл: %v", err)
	}

	t.Run("Тест 1. Инициализация обработчика задач по рассписанию", func(t *testing.T) {
		location, err := time.LoadLocation("Europe/Moscow")
		assert.NoError(t, err)
		year, month, day := time.Now().Date()
		// тест будет выпонятся коректно только если текущее время будет меньше 22:47:03
		// пришлось добавить еще 1 милисекунду что бы корректно работало сравнение
		// при использовании установленного фейкового времени. Если оставить милисекунды
		// равные 0 часто возникае ситуация когда задача выполняется несколько раз что
		// вызывает панику 'panic: sync: negative WaitGroup counter'
		baseTime := time.Date(year, month, day, 22, 47, 3, 1, location)

		fakeClock := clockwork.NewFakeClock()
		td := fakeClock.Until(baseTime)

		sw, err := schedulehandler.NewScheduleHandler(
			schedulehandler.WithTimerJob(conf.GetSchedule().TimerJob),
			schedulehandler.WithDailyJob(conf.GetSchedule().DailyJob),
			schedulehandler.WithFakeClock(fakeClock),
		)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(1)
		sw.Start(
			t.Context(),
			func() error {
				fmt.Println("Start worker 'DailyJob', fakeClock:", sw.ClockWork.Now())
				fmt.Println("Really date:", time.Now())

				assert.True(t, time.Now().Add(td).After(sw.ClockWork.Now()))

				wg.Done()

				return nil
			})

		// смотрим количество job
		assert.Len(t, sw.Scheduler.Jobs(), 1)

		err = fakeClock.BlockUntilContext(context.Background(), 1)
		assert.NoError(t, err)

		fakeClock.Advance(td)

		wg.Wait()

		assert.NoError(t, sw.StopAllJobs())
		assert.NoError(t, sw.Stop())
	})

	t.Run("Тест 2. Инициализация обработчика задач по таймеру", func(t *testing.T) {
		fakeClock := clockwork.NewFakeClock()
		sw, err := schedulehandler.NewScheduleHandler(
			schedulehandler.WithTimerJob(conf.GetSchedule().TimerJob),
			schedulehandler.WithFakeClock(fakeClock),
		)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(1)
		sw.Start(
			t.Context(),
			func() error {
				fmt.Println("Start worker 'DurationJob', fakeClock:", sw.ClockWork.Now())
				fmt.Println("Really date:", time.Now())

				assert.True(t, time.Now().Add(sw.TimerJob).After(sw.ClockWork.Now()))

				wg.Done()

				return nil
			})

		// смотрим количество job
		assert.Len(t, sw.Scheduler.Jobs(), 1)

		err = fakeClock.BlockUntilContext(context.Background(), 1)
		assert.NoError(t, err)

		// запуск по таймеру
		fakeClock.Advance(sw.TimerJob)

		wg.Wait()

		assert.NoError(t, sw.StopAllJobs())
		assert.NoError(t, sw.Stop())
	})

	t.Cleanup(func() {
		os.Unsetenv("GO_ENRICHERZI_MAIN")

		os.Unsetenv("GO_ENRICHERZI_MAIN")
		os.Unsetenv("GO_ENRICHERZI_ZPASSWD")
		os.Unsetenv("GO_ENRICHERZI_NBPASSWD")
		os.Unsetenv("GO_ENRICHERZI_DBWLOGPASSWD")
	})
}
