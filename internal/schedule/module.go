package schedule

import (
	"go.uber.org/fx"

	"github.com/lazerion/outbox-relayer/internal/config"
)

func NewSchedulerProvider(job Job, cfg *config.Config) SchedulerInterface {
	return NewScheduler(job, cfg.Schedule.Interval)
}

var Module = fx.Module(
	"scheduler",
	fx.Provide(
		NewSchedulerProvider,
	),
)
