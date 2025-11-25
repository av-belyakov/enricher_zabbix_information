package schedule

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/subosito/gotenv"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
)

type ScheduleWrapper struct {
	DailyJob  []gocron.AtTime
	TimerJob  time.Duration
	ClockWork *clockwork.FakeClock
}

type ScheduleOptions func(*ScheduleWrapper) error

func NewScheduleHandler(opts ...ScheduleOptions) (*ScheduleWrapper, error) {
	sw := &ScheduleWrapper{
		TimerJob: 60,
	}

	for _, opt := range opts {
		if err := opt(sw); err != nil {
			return sw, err
		}
	}

	return sw, nil
}

// Start запуск обработчика задачи
func (sw *ScheduleWrapper) Start(ctx context.Context, f func() error) error {
	withClock := gocron.WithClock(clockwork.NewRealClock())

	if os.Getenv("GO_ENRICHERZI_MAIN") == "test" && sw.ClockWork != nil {
		withClock = gocron.WithClock(sw.ClockWork)
	}

	s, err := gocron.NewScheduler(withClock)
	if err != nil {
		return err
	}

	go func(ctx context.Context, s gocron.Scheduler) {
		<-ctx.Done()
		s.Shutdown()
	}(ctx, s)

	//используем таймер
	if len(sw.DailyJob) == 0 {
		if _, err := s.NewJob(
			gocron.DurationJob(sw.TimerJob),
			gocron.NewTask(f),
		); err != nil {
			return err
		}
	} else {
		ats := gocron.NewAtTimes(sw.DailyJob[0])
		if len(sw.DailyJob) > 1 {
			ats = gocron.NewAtTimes(sw.DailyJob[0], sw.DailyJob...)
		}

		if _, err = s.NewJob(
			gocron.DailyJob(1, ats),
			gocron.NewTask(f),
		); err != nil {
			return err
		}
	}

	s.Start()

	return nil
}

// WithDailyJob частота запуска задачи в минутах
func WithTimerJob(timerJob int) ScheduleOptions {
	return func(sw *ScheduleWrapper) error {
		if timerJob < 1 {
			return errors.New("the task launch frequency in minutes cannot be less than '1'")
		}

		sw.TimerJob = time.Duration(timerJob) * time.Minute

		return nil
	}
}

// WithDailyJob список времени запуска задачи в формате HH:MM:SS
func WithDailyJob(dailyJob []string) ScheduleOptions {
	return func(sw *ScheduleWrapper) error {
		if len(dailyJob) == 0 {
			return nil
		}

		for _, v := range dailyJob {
			t, err := time.Parse(time.TimeOnly, v)
			if err != nil {
				return errors.New("the time format is incorrect")
			}

			sw.DailyJob = append(sw.DailyJob, gocron.NewAtTime(uint(t.Hour()), uint(t.Minute()), uint(t.Second())))
		}

		return nil
	}
}

// WithFakeClock использование фейкового времени (исключительно для тестирования)
func WithFakeClock(clock *clockwork.FakeClock) ScheduleOptions {
	return func(sw *ScheduleWrapper) error {
		sw.ClockWork = clock

		return nil
	}
}

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
		baseTime := time.Date(year, month, day, 22, 47, 3, 0, location)

		fakeClock := clockwork.NewFakeClock()
		td := fakeClock.Until(baseTime)

		sw, err := NewScheduleHandler(
			WithTimerJob(conf.GetSchedule().TimerJob),
			WithDailyJob(conf.GetSchedule().DailyJob),
			WithFakeClock(fakeClock),
		)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(1)
		sw.Start(t.Context(), func() error {
			fmt.Println("Start worker 'DailyJob', fakeClock:", sw.ClockWork.Now())
			fmt.Println("Really date:", time.Now())

			assert.True(t, time.Now().Add(td).After(sw.ClockWork.Now()))

			wg.Done()

			return nil
		})

		err = fakeClock.BlockUntilContext(context.Background(), 1)
		assert.NoError(t, err)
		fakeClock.Advance(td)

		wg.Wait()
	})

	t.Run("Тест 2. Инициализация обработчика задач по таймеру", func(t *testing.T) {
		fakeClock := clockwork.NewFakeClock()
		sw, err := NewScheduleHandler(
			WithTimerJob(conf.GetSchedule().TimerJob),
			WithFakeClock(fakeClock),
		)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(1)
		sw.Start(t.Context(), func() error {
			fmt.Println("Start worker 'DurationJob', fakeClock:", sw.ClockWork.Now())
			fmt.Println("Really date:", time.Now())

			assert.True(t, time.Now().Add(sw.TimerJob).After(sw.ClockWork.Now()))

			wg.Done()

			return nil
		})

		err = fakeClock.BlockUntilContext(context.Background(), 1)
		assert.NoError(t, err)

		// запуск по таймеру
		fakeClock.Advance(sw.TimerJob)

		wg.Wait()
	})

	t.Cleanup(func() {
		os.Unsetenv("GO_ENRICHERZI_MAIN")

		os.Unsetenv("GO_ENRICHERZI_MAIN")
		os.Unsetenv("GO_ENRICHERZI_ZPASSWD")
		os.Unsetenv("GO_ENRICHERZI_NBPASSWD")
		os.Unsetenv("GO_ENRICHERZI_DBWLOGPASSWD")
	})
}
