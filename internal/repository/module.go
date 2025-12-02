package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lazerion/outbox-relayer/internal/config"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

// NewDB creates a PostgreSQL connection
func NewDB(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to Postgres")
	return db, nil
}

func NewMessageRepositoryProvider(db *sql.DB) MessageRepository {
	return NewPostgresMessageRepository(db)
}

func NewQueryRepositoryProvider(db *sql.DB) QueryRepository {
	return NewPostgresQueryRepository(db)
}

var Module = fx.Module(
	"repository",
	fx.Provide(
		NewDB,
		NewMessageRepositoryProvider,
		NewQueryRepositoryProvider,
	),
)
