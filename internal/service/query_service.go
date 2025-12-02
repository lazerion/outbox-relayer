package service

import (
	"context"
	"fmt"
	"time"

	"github.com/lazerion/outbox-relayer/internal/model"
	"github.com/lazerion/outbox-relayer/internal/repository"
)

type QueryServiceInterface interface {
	ListSentMessages(ctx context.Context, after time.Time, limit int) (*SentMessagesResponse, error)
}

type QueryService struct {
	repo repository.QueryRepository
}

func NewQueryService(repo repository.QueryRepository) QueryServiceInterface {
	return &QueryService{repo: repo}
}

// SentMessagesResponse Cursor-based pagination response
type SentMessagesResponse struct {
	Messages   []model.Message `json:"messages"`
	NextCursor *time.Time      `json:"next_cursor,omitempty"`
}

// ListSentMessages retrieves a page of sent messages after the given cursor
func (s *QueryService) ListSentMessages(ctx context.Context, after time.Time, limit int) (*SentMessagesResponse, error) {
	sentMessages, err := s.repo.ListSentMessages(ctx, after, limit)
	if err != nil {
		return nil, fmt.Errorf("fetch sent messages: %w", err)
	}

	if sentMessages == nil {
		sentMessages = []model.Message{}
	}

	var nextCursor *time.Time
	if len(sentMessages) == limit {
		ts := sentMessages[len(sentMessages)-1].SentTime
		nextCursor = &ts
	}

	return &SentMessagesResponse{
		Messages:   sentMessages,
		NextCursor: nextCursor,
	}, nil
}
