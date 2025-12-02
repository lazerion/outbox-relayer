package schedule

import (
	"context"
	"log"

	"go.uber.org/fx"
)

// StartStopSchedulerHook starts the scheduler on Fx startup and stops on shutdown
func StartStopSchedulerHook(lc fx.Lifecycle, sched SchedulerInterface) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Starting scheduler...")
			go sched.Start(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Stopping scheduler...")
			sched.Stop()
			return nil
		},
	})
}

var ModuleWithLifeCycle = fx.Module(
	"scheduler-lifecycle",
	fx.Invoke(StartStopSchedulerHook),
)
