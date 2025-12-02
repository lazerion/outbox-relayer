package infra

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lazerion/outbox-relayer/internal/config"
	"go.uber.org/fx"
)

func RunMigrations(cfg *config.Config, db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(cfg.Migration.Path)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+absPath,
		"postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Println("Migrations applied successfully")
	return nil
}

var Module = fx.Module(
	"migrations",
	fx.Invoke(func(lc fx.Lifecycle, cfg *config.Config, db *sql.DB) {
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				return RunMigrations(cfg, db)
			},
			OnStop: func(_ context.Context) error { return nil },
		})
	}),
)
