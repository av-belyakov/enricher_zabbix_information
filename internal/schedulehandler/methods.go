package schedulehandler

import (
	"context"
	"os"

	"github.com/go-co-op/gocron/v2"
	"github.com/jonboulle/clockwork"
)

// Start запуск обработчика расписания
func (sw *ScheduleWrapper) Start(ctx context.Context, f func() error) error {
	withClock := gocron.WithClock(clockwork.NewRealClock())

	// для тестов будет использоватся фейковое время
	if os.Getenv("GO_ENRICHERZI_MAIN") == "test" && sw.ClockWork != nil {
		withClock = gocron.WithClock(sw.ClockWork)
	}

	s, err := gocron.NewScheduler(withClock)
	if err != nil {
		return err
	}

	sw.Scheduler = s

	go func(ctx context.Context, s gocron.Scheduler) {
		<-ctx.Done()
		s.Shutdown()
	}(ctx, s)

	//используем таймер
	if len(sw.DailyJob) == 0 {
		if _, err := s.NewJob(
			gocron.DurationJob(sw.TimerJob),
			gocron.NewTask(f),
			gocron.WithName(Worker_Name),
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
			gocron.WithName(Worker_Name),
		); err != nil {
			return err
		}
	}

	s.Start()

	return nil
}

// Stop остановка обработчика расписания
func (sw *ScheduleWrapper) Stop() error {
	if sw.Scheduler != nil {
		return sw.Scheduler.Shutdown()
	}

	return nil
}

// StopAllJobs остановка всех задач в расписании
func (sw *ScheduleWrapper) StopAllJobs() error {
	if sw.Scheduler != nil {
		return sw.Scheduler.StopJobs()
	}

	return nil
}
