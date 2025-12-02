package handler

import (
	"net/http"

	"github.com/lazerion/outbox-relayer/internal/schedule"
)

type StatusResponse struct {
	Status string `json:"status"`
}

// SchedulerHandler manages toggling of a background scheduler
type SchedulerHandler struct {
	sched schedule.SchedulerInterface
}

// NewSchedulerHandler creates a new handler
func NewSchedulerHandler(s schedule.SchedulerInterface) *SchedulerHandler {
	return &SchedulerHandler{sched: s}
}

// ToggleScheduler godoc
// @Summary Toggle the message scheduler
// @Description Starts the scheduler if it is stopped, or stops it if it is running.
// @Tags Scheduler
// @Produce json
// @Success 200 {object} handler.StatusResponse "Scheduler started or stopped"
// @Router /api/v1/scheduler/toggle [post]
func (h *SchedulerHandler) ToggleScheduler(w http.ResponseWriter, r *http.Request) {
	if running := h.sched.IsRunning(); !running {
		h.sched.Start(r.Context())
		WriteJSON(w, http.StatusOK, StatusResponse{Status: "Scheduler started"})
	} else {
		h.sched.Stop()
		WriteJSON(w, http.StatusOK, StatusResponse{Status: "Scheduler stopped"})
	}
}
