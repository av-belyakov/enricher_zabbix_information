package schedulehandler

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jonboulle/clockwork"
)

// ScheduleWrapper обертка для gocron.Scheduler
type ScheduleWrapper struct {
	Scheduler gocron.Scheduler
	DailyJob  []gocron.AtTime
	TimerJob  time.Duration
	ClockWork *clockwork.FakeClock
}

// ScheduleOptions опции для ScheduleWrapper
type ScheduleOptions func(*ScheduleWrapper) error
