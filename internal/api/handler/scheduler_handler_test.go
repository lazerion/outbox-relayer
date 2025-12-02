package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/lazerion/outbox-relayer/internal/api/handler"
)

// MockScheduler implements SchedulerInterface for testing
type MockScheduler struct {
	startCalled  int32
	stopCalled   int32
	runningState bool
}

func (m *MockScheduler) Start(parentCtx context.Context) {
	atomic.AddInt32(&m.startCalled, 1)
	m.runningState = true
}

func (m *MockScheduler) Stop() {
	atomic.AddInt32(&m.stopCalled, 1)
	m.runningState = false
}

func (m *MockScheduler) IsRunning() bool {
	return m.runningState
}

func TestToggleScheduler_Multiple(t *testing.T) {
	mock := &MockScheduler{}
	h := handler.NewSchedulerHandler(mock)

	tests := []struct {
		name         string
		expectBody   string
		expectStart  int32
		expectStop   int32
		runningAfter bool
	}{
		{"Start scheduler", "Scheduler started", 1, 0, true},
		{"Stop scheduler", "Scheduler stopped", 1, 1, false},
		{"Start again", "Scheduler started", 2, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/scheduler/toggle", nil)
			w := httptest.NewRecorder() // reset recorder for each subtest

			h.ToggleScheduler(w, req)

			resp := w.Result()
			body := w.Body.String()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status 200, got %d", resp.StatusCode)
			}

			if !strings.Contains(body, tt.expectBody) {
				t.Errorf("expected body to contain %q, got %s", tt.expectBody, body)
			}

			if atomic.LoadInt32(&mock.startCalled) != tt.expectStart {
				t.Errorf("expected Start called %d times, got %d", tt.expectStart, mock.startCalled)
			}

			if atomic.LoadInt32(&mock.stopCalled) != tt.expectStop {
				t.Errorf("expected Stop called %d times, got %d", tt.expectStop, mock.stopCalled)
			}

			if mock.IsRunning() != tt.runningAfter {
				t.Errorf("expected running=%v, got %v", tt.runningAfter, mock.IsRunning())
			}
		})
	}
}
