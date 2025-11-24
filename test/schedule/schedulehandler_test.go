package schedule

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/stretchr/testify/assert"
	"github.com/subosito/gotenv"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
)

type ScheduleWrapper struct {
	DailyJob []string
	TimerJob time.Duration
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

	/*
		if len(sw.DailyJob) > 0 {
			var dailyJobs []gocron.AtTime
			for _, v := range sw.DailyJob {
				t, err := time.Parse(time.TimeOnly, v)
				if err != nil {
					return sw, errors.New("the time format is incorrect")
				}

				fmt.Printf("Only test:\nhour:'%d',\nminute:'%d',\nsecond:'%d'\n", t.Hour(), t.Minute(), t.Second())

				//dailyJobs = append(dailyJobs, gocron.NewAtTime(uint(hour), uint(minute), uint(second)))
				dailyJobs = append(dailyJobs, gocron.NewAtTime(uint(t.Hour()), uint(t.Minute()), uint(t.Second())))
			}
		}
	*/

	return sw, nil
}

// Start запуск обработчика задачи
func (sw *ScheduleWrapper) Start(ctx context.Context, f func() error) error {
	s, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	go func(ctx context.Context, s gocron.Scheduler) {
		<-ctx.Done()
		s.Shutdown()
	}(ctx, s)

	//используем таймер
	if len(sw.DailyJob) == 0 {
		_, err := s.NewJob(
			gocron.DurationJob(time.Minute*sw.TimerJob),
			gocron.NewTask(f),
		)
		if err != nil {
			return err
		}

		s.Start()

		return nil
	}

	var dailyJobs []gocron.AtTime
	for _, v := range sw.DailyJob {
		t, err := time.Parse(time.TimeOnly, v)
		if err != nil {
			return errors.New("the time format is incorrect")
		}

		dailyJobs = append(dailyJobs, gocron.NewAtTime(uint(t.Hour()), uint(t.Minute()), uint(t.Second())))
	}

	ats := gocron.NewAtTimes(dailyJobs[0])
	if len(dailyJobs) > 1 {
		ats = gocron.NewAtTimes(dailyJobs[0], dailyJobs...)
	}

	_, err = s.NewJob(
		gocron.DailyJob(1, ats),
		gocron.NewTask(f),
	)
	if err != nil {
		return err
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

// WithDailyJob список времени запуска задачи в формате HH:MM
func WithDailyJob(dailyJob []string) ScheduleOptions {
	return func(sw *ScheduleWrapper) error {
		sw.DailyJob = dailyJob
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

	t.Run("Тест 1. Инициализация обработчика задач по рассписанию или по таймеру", func(t *testing.T) {
		sw, err := NewScheduleHandler(
			WithTimerJob(conf.Schedule.TimerJob),
			WithDailyJob(conf.GetSchedule().DailyJob),
		)
		assert.NoError(t, err)

		fmt.Println("ScheduleWrapper:", sw)
	})

	t.Cleanup(func() {
		os.Unsetenv("GO_ENRICHERZI_MAIN")
		os.Unsetenv("GO_ENRICHERZI_ZPASSWD")
		os.Unsetenv("GO_ENRICHERZI_NBPASSWD")
		os.Unsetenv("GO_ENRICHERZI_DBWLOGPASSWD")
	})
}
