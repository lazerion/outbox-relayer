package schedule_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lazerion/outbox-relayer/internal/schedule"
	"github.com/stretchr/testify/assert"
)

type MockJob struct {
	runCount int32
	fail     bool
	delay    time.Duration
	started  chan struct{}
}

func (m *MockJob) Run(ctx context.Context) error {
	if m.started != nil {
		m.started <- struct{}{}
	}
	atomic.AddInt32(&m.runCount, 1)
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.fail {
		return errors.New("job failed")
	}
	return nil
}

func TestScheduler_RunOnceImmediately(t *testing.T) {
	job := &MockJob{}
	s := schedule.NewScheduler(job, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	count := atomic.LoadInt32(&job.runCount)
	assert.GreaterOrEqual(t, count, int32(1), "scheduler should run immediately on start")
	s.Stop()
}

func TestScheduler_NoOverlap(t *testing.T) {
	const jobDelay = 30 * time.Millisecond
	const tickInterval = 10 * time.Millisecond

	job := &MockJob{delay: jobDelay}
	s := schedule.NewScheduler(job, tickInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	s.Stop()

	count := atomic.LoadInt32(&job.runCount)
	maxExpectedRuns := int32((100*time.Millisecond)/jobDelay) + 1
	assert.True(t, count <= maxExpectedRuns, "job should not overlap")
}

func TestScheduler_ContinuesOnFailure(t *testing.T) {
	const tickInterval = 20 * time.Millisecond
	const testDuration = 100 * time.Millisecond

	job := &MockJob{fail: true}
	s := schedule.NewScheduler(job, tickInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)
	time.Sleep(testDuration)
	s.Stop()

	count := atomic.LoadInt32(&job.runCount)
	expectedMinRuns := int32(testDuration / tickInterval)
	assert.GreaterOrEqual(t, count, expectedMinRuns, "scheduler should continue next cycle despite failures")
}

func TestScheduler_StopWaitsForInFlight(t *testing.T) {
	const jobDelay = 50 * time.Millisecond
	started := make(chan struct{}, 1)
	job := &MockJob{delay: jobDelay, started: started}
	s := schedule.NewScheduler(job, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)

	// wait until job actually starts
	<-started

	start := time.Now()
	s.Stop()
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed.Milliseconds(), jobDelay.Milliseconds(), "Stop should wait for in-flight job to finish")
}

// Test multiple Start calls do not create multiple schedulers
func TestScheduler_MultipleStart(t *testing.T) {
	const tickInterval = 20 * time.Millisecond
	job := &MockJob{}
	s := schedule.NewScheduler(job, tickInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)
	s.Start(ctx) // second call should not panic or start another loop
	time.Sleep(50 * time.Millisecond)
	s.Stop()

	count := atomic.LoadInt32(&job.runCount)
	assert.GreaterOrEqual(t, count, int32(1), "scheduler should have run at least once")
	assert.False(t, s.IsRunning())
}
