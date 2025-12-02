package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lazerion/outbox-relayer/internal/gateway"
	"github.com/lazerion/outbox-relayer/internal/repository"
	"github.com/lazerion/outbox-relayer/internal/schedule"
)

type SentMessageEvent struct {
	MessageID string
	SentAt    time.Time
}

type RelayerService struct {
	repo        repository.MessageRepository
	sender      gateway.Sender
	batch       int
	timeout     time.Duration
	maxAttempts int
	cacheCh     chan SentMessageEvent
}

func NewRelayerService(repo repository.MessageRepository, sender gateway.Sender, batch int, timeout time.Duration, maxAttempts int,
	cacheCh chan SentMessageEvent) schedule.Job {
	return &RelayerService{
		repo:        repo,
		sender:      sender,
		batch:       batch,
		timeout:     timeout,
		maxAttempts: maxAttempts,
		cacheCh:     cacheCh,
	}
}

// Run fetches pending messages and sends them with retry/attempt logic
// Transactional safety is ensured by wrapping all pending message updates in a single database transaction (`tx`).
// Each message is marked sent, failed, or attempt incremented atomically.
func (s *RelayerService) Run(ctx context.Context) error {
	msgs, tx, err := s.repo.FetchPendingTx(ctx, s.batch)
	if err != nil {
		return fmt.Errorf("fetch pending messages: %w", err)
	}

	if len(msgs) == 0 {
		_ = tx.Rollback()
		return nil
	}

	for _, m := range msgs {
		if m.AttemptCount >= s.maxAttempts {
			log.Printf("message ID %d exceeded max attempts (%d), marking as failed", m.ID, s.maxAttempts)
			_ = s.repo.MarkAsFailedTx(ctx, tx, m.ID)
			continue
		}

		sendCtx, cancel := context.WithTimeout(ctx, s.timeout)
		resp, err := s.sender.Send(sendCtx, m)
		cancel()

		if err != nil {
			if gateway.IsRecoverable(err) {
				log.Printf("recoverable error sending message ID %d: %v", m.ID, err)
				_ = s.repo.IncrementAttemptTx(ctx, tx, m.ID)
			} else {
				log.Printf("unrecoverable error sending message ID %d: %v", m.ID, err)
				_ = s.repo.MarkAsFailedTx(ctx, tx, m.ID)
			}
			continue
		}

		switch strings.ToLower(resp.Message) {
		case "accepted":
			now := time.Now()
			if err := s.repo.MarkAsSentTx(ctx, tx, m.ID, resp.MessageID, now); err != nil {
				log.Printf("failed to mark message ID %d as sent: %v", m.ID, err)
				continue
			}
			// Push to cache channel asynchronously, non-blocking
			select {
			case s.cacheCh <- SentMessageEvent{MessageID: resp.MessageID, SentAt: now}:
			default:
				log.Printf("cache channel full, skipping caching for message ID %d", m.ID)
			}

		default:
			log.Printf("sender rejected message ID %d, marking failed: status=%s",
				m.ID, resp.Message)
			_ = s.repo.MarkAsFailedTx(ctx, tx, m.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}
