package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lazerion/outbox-relayer/internal/model"
	"github.com/lazerion/outbox-relayer/internal/service"
)

// Mock repository
type MockMessageRepo struct {
	Messages []model.Message
	Err      error
}

func (m *MockMessageRepo) ListSentMessages(ctx context.Context, after time.Time, limit int) ([]model.Message, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Messages, nil
}

func TestQueryService_ListSentMessages(t *testing.T) {
	msgTime := time.Now()
	mockRepo := &MockMessageRepo{
		Messages: []model.Message{
			{ID: 1, PhoneNumber: "+123", Content: "Hello", SentTime: msgTime},
			{ID: 2, PhoneNumber: "+456", Content: "World", SentTime: msgTime.Add(time.Minute)},
		},
	}

	svc := service.NewQueryService(mockRepo)
	ctx := context.Background()

	resp, err := svc.ListSentMessages(ctx, time.Time{}, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(resp.Messages))
	}

	if resp.NextCursor == nil || !resp.NextCursor.Equal(mockRepo.Messages[1].SentTime) {
		t.Errorf("expected NextCursor=%v, got %v", mockRepo.Messages[1].SentTime, resp.NextCursor)
	}
}

func TestQueryService_ListSentMessages_Error(t *testing.T) {
	mockRepo := &MockMessageRepo{
		Err: errors.New("db error"),
	}

	svc := service.NewQueryService(mockRepo)
	_, err := svc.ListSentMessages(context.Background(), time.Time{}, 2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "fetch sent messages: db error" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestQueryService_ListSentMessages_Empty(t *testing.T) {
	mockRepo := &MockMessageRepo{
		Messages: []model.Message{},
	}

	svc := service.NewQueryService(mockRepo)
	resp, err := svc.ListSentMessages(context.Background(), time.Time{}, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(resp.Messages))
	}

	if resp.NextCursor != nil {
		t.Errorf("expected NextCursor=nil, got %v", resp.NextCursor)
	}
}
