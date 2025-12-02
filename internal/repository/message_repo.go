package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lazerion/outbox-relayer/internal/model"
)

type MessageRepository interface {
	FetchPendingTx(ctx context.Context, batchSize int) ([]model.Message, *sql.Tx, error)
	MarkAsSentTx(ctx context.Context, tx *sql.Tx, id int64, externalID string, sentTime time.Time) error
	MarkAsFailedTx(ctx context.Context, tx *sql.Tx, id int64) error
	IncrementAttemptTx(ctx context.Context, tx *sql.Tx, id int64) error
}

type PostgresMessageRepository struct {
	db *sql.DB
}

func NewPostgresMessageRepository(db *sql.DB) MessageRepository {
	return &PostgresMessageRepository{db: db}
}

// FetchPendingTx
// utilizing `FOR UPDATE SKIP LOCKED` to prevent race conditions and ensure reliability in a multi-instance environment
func (r *PostgresMessageRepository) FetchPendingTx(ctx context.Context, batchSize int) ([]model.Message, *sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return nil, nil, err
	}

	rows, err := tx.QueryContext(ctx,
		`SELECT id, phone_number, content, status 
         FROM messages 
         WHERE status = 'pending' 
         ORDER BY id 
         LIMIT $1 
         FOR UPDATE SKIP LOCKED`,
		batchSize,
	)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	defer rows.Close()

	var msgs []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.PhoneNumber, &m.Content, &m.Status); err != nil {
			tx.Rollback()
			return nil, nil, err
		}
		msgs = append(msgs, m)
	}

	return msgs, tx, nil
}

func (r *PostgresMessageRepository) IncrementAttemptTx(ctx context.Context, tx *sql.Tx, id int64) error {
	_, err := tx.ExecContext(ctx, `
        UPDATE messages
        SET attempt_count = attempt_count + 1
        WHERE id = $1
    `, id)
	return err
}

func (r *PostgresMessageRepository) MarkAsSentTx(
	ctx context.Context,
	tx *sql.Tx,
	id int64,
	externalID string,
	sentTime time.Time,
) error {
	_, err := tx.ExecContext(ctx, `
        UPDATE messages
        SET status = 'sent',
            external_id = $2,
            sent_time = $3
        WHERE id = $1
    `, id, externalID, sentTime)
	return err
}

func (r *PostgresMessageRepository) MarkAsFailedTx(ctx context.Context, tx *sql.Tx, id int64) error {
	_, err := tx.ExecContext(ctx, `UPDATE messages SET status='failed' WHERE id=$1`, id)
	return err
}
