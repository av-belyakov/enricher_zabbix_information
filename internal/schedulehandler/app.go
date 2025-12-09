package schedulehandler

import (
	"errors"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jonboulle/clockwork"
)

// NewScheduleHandler новый обработчик расписания
func NewScheduleHandler(opts ...ScheduleOptions) (*ScheduleWrapper, error) {
	sw := &ScheduleWrapper{
		TimerJob: Default_Timer_Job,
	}

	for _, opt := range opts {
		if err := opt(sw); err != nil {
			return sw, err
		}
	}

	return sw, nil
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
