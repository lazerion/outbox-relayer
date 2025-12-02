package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lazerion/outbox-relayer/internal/service"
)

type QueryHandler struct {
	service service.QueryServiceInterface
}

func NewQueryHandler(qs service.QueryServiceInterface) *QueryHandler {
	return &QueryHandler{service: qs}
}

// ListSentMessages retrieves sent SMS messages using cursor-based pagination.
//
// @Summary      List sent messages
// @Description  Returns a list of messages with status `sent`, ordered by `sent_time` ascending.
// @Description  Supports cursor-based pagination via the `after` cursor.
// @Tags         messages
// @Accept       json
// @Produce      json
//
// @Param        after  query     string  false  "Return messages sent after this timestamp (RFC3339)"
// @Param        limit  query     int     false  "Number of messages to return (1–50), defaults to 42"
//
// @Success      200  {array}  service.SentMessagesResponse
// @Failure      400  {object}  ErrorResponse  "Invalid request format"
// @Failure      500  {object}  ErrorResponse  "Internal server error"
//
// @Router       /api/v1/messages/sent [get]
func (h *QueryHandler) ListSentMessages(w http.ResponseWriter, r *http.Request) {
	const (
		defaultLimit = 42
		maxLimit     = 50
	)

	ctx := r.Context()
	q := r.URL.Query()

	var after time.Time
	if v := q.Get("after"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid 'after' timestamp")
			return
		}
		after = t
	}

	limit := defaultLimit
	if v := q.Get("limit"); v != "" {
		l, err := strconv.Atoi(v)
		if err != nil || l <= 0 {
			WriteError(w, http.StatusBadRequest, "'limit' must be a positive integer")
			return
		}
		limit = l
	}

	if limit > maxLimit {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("'limit' cannot exceed %d", maxLimit))
		return
	}

	resp, err := h.service.ListSentMessages(ctx, after, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}
