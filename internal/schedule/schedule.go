package schedule

import (
	"context"
	"log"
	"sync"
	"time"
)

type Job interface {
	Run(ctx context.Context) error
}

type SchedulerInterface interface {
	Start(parentCtx context.Context)
	Stop()
	IsRunning() bool
}

type Scheduler struct {
	job      Job
	interval time.Duration

	mu      sync.Mutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wgJob   sync.WaitGroup
	wgMain  sync.WaitGroup
}

func NewScheduler(job Job, interval time.Duration) SchedulerInterface {
	return &Scheduler{
		job:      job,
		interval: interval,
	}
}

// Start begins periodic execution of the job in a non-blocking way
func (s *Scheduler) Start(parentCtx context.Context) {
	s.mu.Lock()
	if s.ctx != nil {
		s.mu.Unlock()
		log.Println("scheduler already started")
		return
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	ctx := s.ctx
	s.wgMain.Add(1)
	s.mu.Unlock()

	go func() {
		defer s.wgMain.Done()

		s.runOnce(ctx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Context cancelled, wait for any in-flight job to finish before exiting
				s.wgJob.Wait()
				log.Println("scheduler stopped gracefully")
				return
			case <-ticker.C:
				s.runOnce(ctx)
			}
		}
	}()
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
		s.ctx = nil
	}
	s.mu.Unlock()
	s.wgMain.Wait()
}

// IsRunning checks the state safely
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ctx != nil
}

// runOnce ensures no overlapping job executions
func (s *Scheduler) runOnce(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		log.Println("job already running, skipping this tick")
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wgJob.Add(1)
	go func() {
		defer func() {
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
			s.wgJob.Done()
		}()

		if err := s.job.Run(ctx); err != nil {
			log.Println("job error:", err)
		}
	}()
}
