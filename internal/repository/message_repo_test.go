package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazerion/outbox-relayer/internal/repository"
)

func TestPostgresMessageRepository_FetchPendingTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	repo := repository.NewPostgresMessageRepository(db)

	mock.ExpectBegin()

	msgsRows := sqlmock.NewRows([]string{"id", "phone_number", "content", "status"}).
		AddRow(int64(1), "+123456789", "Hello", "pending").
		AddRow(int64(2), "+987654321", "World", "pending")

	mock.ExpectQuery(`SELECT id, phone_number, content, status`).
		WithArgs(2).
		WillReturnRows(msgsRows)

	ctx := context.Background()
	msgs, tx, err := repo.FetchPendingTx(ctx, 2)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if tx == nil {
		t.Fatal("expected transaction, got nil")
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].ID != 1 || msgs[1].ID != 2 {
		t.Errorf("unexpected message IDs: %v, %v", msgs[0].ID, msgs[1].ID)
	}

	tx.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresMessageRepository_MarkAsSentTx(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := repository.NewPostgresMessageRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE messages`).
		WithArgs(int64(1), "ext123", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, _ := db.BeginTx(context.Background(), nil)
	err := repo.MarkAsSentTx(context.Background(), tx, 1, "ext123", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	tx.Commit()
}

func TestPostgresMessageRepository_MarkAsFailedTx(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := repository.NewPostgresMessageRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE messages SET status='failed'`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, _ := db.BeginTx(context.Background(), nil)
	err := repo.MarkAsFailedTx(context.Background(), tx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	tx.Commit()
}

func TestPostgresMessageRepository_IncrementAttemptTx(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := repository.NewPostgresMessageRepository(db)
	ctx := context.Background()
	messageID := int64(42)

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("db.BeginTx failed: %v", err)
	}
	mock.ExpectExec(`UPDATE messages`).
		WithArgs(messageID).
		WillReturnResult(sqlmock.NewResult(0, 1)) // Expect 1 row affected

	err = repo.IncrementAttemptTx(ctx, tx, messageID)
	if err != nil {
		t.Errorf("expected no error from IncrementAttemptTx, got: %v", err)
	}

	tx.Commit()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
