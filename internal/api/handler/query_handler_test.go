package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lazerion/outbox-relayer/internal/api/handler"
	"github.com/lazerion/outbox-relayer/internal/model"
	"github.com/lazerion/outbox-relayer/internal/service"
)

type MockQueryService struct {
	resp *service.SentMessagesResponse
	err  error
}

func (m *MockQueryService) ListSentMessages(ctx context.Context, after time.Time, limit int) (*service.SentMessagesResponse, error) {
	return m.resp, m.err
}

func TestListSentMessages(t *testing.T) {
	msgTime := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	nextCursor := msgTime.Add(time.Minute)

	sentResp := &service.SentMessagesResponse{
		Messages: []model.Message{
			{ID: 1, Content: "Hello", SentTime: msgTime},
		},
		NextCursor: &nextCursor,
	}

	tests := []struct {
		name           string
		query          string
		mockResp       *service.SentMessagesResponse
		mockErr        error
		wantStatusCode int
		wantBodySubstr string
	}{
		{
			name:           "Valid request",
			query:          "limit=1",
			mockResp:       sentResp,
			mockErr:        nil,
			wantStatusCode: http.StatusOK,
			wantBodySubstr: `"id":1`,
		},
		{
			name:           "Missing limit",
			query:          "",
			mockResp:       sentResp,
			mockErr:        nil,
			wantStatusCode: http.StatusOK,
			wantBodySubstr: `"id":1`,
		},
		{
			name:           "Invalid limit (negative)",
			query:          "limit=-1",
			mockResp:       nil,
			mockErr:        nil,
			wantStatusCode: http.StatusBadRequest,
			wantBodySubstr: "'limit' must be a positive integer",
		},
		{
			name:           "Limit exceeds max",
			query:          "limit=100",
			mockResp:       nil,
			mockErr:        nil,
			wantStatusCode: http.StatusBadRequest,
			wantBodySubstr: "'limit' cannot exceed",
		},
		{
			name:           "Invalid after timestamp",
			query:          "limit=1&after=invalid",
			mockResp:       nil,
			mockErr:        nil,
			wantStatusCode: http.StatusBadRequest,
			wantBodySubstr: "invalid 'after' timestamp",
		},
		{
			name:           "Service error",
			query:          "limit=1",
			mockResp:       nil,
			mockErr:        errors.New("service failed"),
			wantStatusCode: http.StatusInternalServerError,
			wantBodySubstr: "service failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockQueryService{
				resp: tt.mockResp,
				err:  tt.mockErr,
			}

			h := handler.NewQueryHandler(mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/messages/sent?"+tt.query, nil)
			w := httptest.NewRecorder()

			h.ListSentMessages(w, req)

			resp := w.Result()
			body := w.Body.String()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status %d, got %d", tt.wantStatusCode, resp.StatusCode)
			}

			if !strings.Contains(body, tt.wantBodySubstr) {
				t.Errorf("expected body to contain %q, got %s", tt.wantBodySubstr, body)
			}
		})
	}
}

func TestListSentMessages_NextCursorIncluded(t *testing.T) {
	msgTime := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	nextCursor := msgTime.Add(time.Minute)

	mockResp := &service.SentMessagesResponse{
		Messages: []model.Message{
			{ID: 1, Content: "Hello", SentTime: msgTime},
		},
		NextCursor: &nextCursor,
	}

	mockSvc := &MockQueryService{
		resp: mockResp,
	}

	h := handler.NewQueryHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/messages/sent?limit=1", nil)
	w := httptest.NewRecorder()

	h.ListSentMessages(w, req)

	resp := w.Result()
	body := w.Body.String()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if !strings.Contains(body, `"id":1`) {
		t.Errorf("expected body to contain message ID, got %s", body)
	}

	expectedCursor := nextCursor.Format(time.RFC3339)
	if !strings.Contains(body, expectedCursor) {
		t.Errorf("expected body to contain NextCursor %q, got %s", expectedCursor, body)
	}
}
