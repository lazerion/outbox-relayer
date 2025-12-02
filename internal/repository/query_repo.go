package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lazerion/outbox-relayer/internal/model"
)

type QueryRepository interface {
	ListSentMessages(ctx context.Context, after time.Time, limit int) ([]model.Message, error)
}

type PostgresQueryRepository struct {
	db *sql.DB
}

func NewPostgresQueryRepository(db *sql.DB) QueryRepository {
	return &PostgresQueryRepository{db: db}
}

func (r *PostgresQueryRepository) ListSentMessages(ctx context.Context, after time.Time, limit int) ([]model.Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, phone_number, content, status, sent_time, external_id
		FROM messages
		WHERE status = 'sent' AND sent_time > $1
		ORDER BY sent_time ASC
		LIMIT $2
	`, after, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.PhoneNumber, &m.Content, &m.Status, &m.SentTime, &m.ExternalID); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}

	return msgs, nil
}
