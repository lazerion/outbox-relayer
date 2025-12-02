package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazerion/outbox-relayer/internal/repository"
)

func TestPostgresQueryRepository_ListSentMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	repo := repository.NewPostgresQueryRepository(db)

	msgTime := time.Now()
	rows := sqlmock.NewRows([]string{"id", "phone_number", "content", "status", "sent_time", "external_id"}).
		AddRow("1", "+123456789", "Hello", "sent", msgTime, "ext1").
		AddRow("2", "+987654321", "World", "sent", msgTime.Add(time.Minute), "ext2")

	// Expect query
	mock.ExpectQuery(`SELECT id, phone_number, content, status, sent_time, external_id`).
		WithArgs(sqlmock.AnyArg(), 2).
		WillReturnRows(rows)

	msgs, err := repo.ListSentMessages(context.Background(), time.Time{}, 2)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}

	if msgs[0].ID != 1 || msgs[1].ID != 2 {
		t.Errorf("unexpected message IDs: %v, %v", msgs[0].ID, msgs[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
