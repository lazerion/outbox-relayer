package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazerion/outbox-relayer/internal/gateway"
	"github.com/lazerion/outbox-relayer/internal/model"
	"github.com/lazerion/outbox-relayer/internal/service"
	"github.com/stretchr/testify/require"
)

type MockMessageRepository struct {
	FetchPendingTxFunc func(ctx context.Context, batchSize int) ([]model.Message, *sql.Tx, error)
}

func (m *MockMessageRepository) FetchPendingTx(ctx context.Context, batchSize int) ([]model.Message, *sql.Tx, error) {
	return m.FetchPendingTxFunc(ctx, batchSize)
}
func (m *MockMessageRepository) MarkAsSentTx(ctx context.Context, tx *sql.Tx, id int64, messageID string, sentAt time.Time) error {
	return nil
}
func (m *MockMessageRepository) MarkAsFailedTx(ctx context.Context, tx *sql.Tx, id int64) error {
	return nil
}
func (m *MockMessageRepository) IncrementAttemptTx(ctx context.Context, tx *sql.Tx, id int64) error {
	return nil
}

type mockSender struct{}

func (m *mockSender) Send(ctx context.Context, msg model.Message) (*gateway.SendResponse, error) {
	return &gateway.SendResponse{
		MessageID: "1",
		Message:   "accepted",
	}, nil
}

func TestRelayerService_Run(t *testing.T) {
	tests := []struct {
		name           string
		pendingMsgs    []model.Message
		expectCommit   bool
		expectRollback bool
		expectCacheEvt bool
	}{
		{
			name: "with pending message",
			pendingMsgs: []model.Message{
				{ID: 123, PhoneNumber: "+123456789", Content: "hello", AttemptCount: 0},
			},
			expectCommit:   true,
			expectRollback: false,
			expectCacheEvt: true,
		},
		{
			name:           "no pending messages",
			pendingMsgs:    []model.Message{},
			expectCommit:   false,
			expectRollback: true,
			expectCacheEvt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectBegin()
			tx, err := db.Begin()
			require.NoError(t, err)

			if tt.expectCommit {
				mock.ExpectCommit()
			}
			if tt.expectRollback {
				mock.ExpectRollback()
			}

			repo := &MockMessageRepository{
				FetchPendingTxFunc: func(ctx context.Context, batchSize int) ([]model.Message, *sql.Tx, error) {
					return tt.pendingMsgs, tx, nil
				},
			}
			cacheChan := make(chan service.SentMessageEvent, 1)
			relayer := service.NewRelayerService(repo, &mockSender{}, 10, time.Second, 3, cacheChan)
			err = relayer.Run(context.Background())
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())

			if tt.expectCacheEvt {
				select {
				case evt := <-cacheChan:
					require.Equal(t, "1", evt.MessageID)
				default:
					t.Fatalf("expected cache event but none received")
				}
			} else {
				select {
				case <-cacheChan:
					t.Fatalf("did not expect cache event but received one")
				default:
					// ok
				}
			}
		})
	}
}
