package schedule_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lazerion/outbox-relayer/internal/schedule"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// mockScheduler implements schedule.SchedulerInterface
type mockScheduler struct {
	started int32
	stopped int32
}

func (m *mockScheduler) Start(ctx context.Context) { atomic.StoreInt32(&m.started, 1) }
func (m *mockScheduler) Stop()                     { atomic.StoreInt32(&m.stopped, 1) }
func (m *mockScheduler) IsRunning() bool           { return atomic.LoadInt32(&m.started) == 1 }

func TestStartStopSchedulerHook(t *testing.T) {
	mockSched := &mockScheduler{}

	app := fx.New(
		fx.Provide(func() schedule.SchedulerInterface {
			return mockSched
		}),
		fx.Invoke(schedule.StartStopSchedulerHook),
	)

	require.NoError(t, app.Start(context.Background()))
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&mockSched.started) == 1
	}, time.Second, 10*time.Millisecond, "scheduler did not start")

	require.NoError(t, app.Stop(context.Background()))
	require.Equal(t, int32(1), atomic.LoadInt32(&mockSched.stopped))
}
