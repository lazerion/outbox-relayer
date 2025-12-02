package integration

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/lazerion/outbox-relayer/internal/config"
	"github.com/lazerion/outbox-relayer/internal/gateway"
	"github.com/lazerion/outbox-relayer/internal/infra"
	"github.com/lazerion/outbox-relayer/internal/repository"
	"github.com/lazerion/outbox-relayer/internal/schedule"
	"github.com/lazerion/outbox-relayer/internal/service"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/fx"
)

func setupPostgresContainer(t *testing.T) (testcontainers.Container, *sql.DB) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	pg, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	host, _ := pg.Host(ctx)
	port, _ := pg.MappedPort(ctx, "5432")
	dsn := "host=" + host + " port=" + port.Port() + " user=test password=test dbname=testdb sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	return pg, db
}

func TestRelayerIntegration(t *testing.T) {
	ctx := context.Background()

	pg, db := setupPostgresContainer(t)
	defer pg.Terminate(ctx)
	defer db.Close()

	app := fx.New(
		fx.Provide(func() *config.Config {
			host, _ := pg.Host(ctx)
			port, _ := pg.MappedPort(ctx, "5432")

			return &config.Config{
				Postgres: config.PostgresConfig{
					Host:     host,
					Port:     port.Int(), // <-- mapped port
					User:     "test",
					Password: "test",
					Database: "testdb",
				},
				Relayer: config.RelayerConfig{
					Batch:       10,
					Timeout:     time.Second,
					MaxAttempts: 3,
				},
				Webhook: config.WebhookConfig{
					Url:     "https://webhook.site/5dcc9ebe-72f4-4332-bff0-94569bf9748e",
					AuthKey: "",
					Timeout: time.Second,
				},
				Schedule: config.ScheduleConfig{
					Interval: 2 * time.Second,
				},
				Migration: config.Migration{
					Path: "../../infra/migrations",
				},
			}
		}),
		infra.Module,
		repository.Module,
		gateway.Module,
		service.Module,
		schedule.Module,
		schedule.ModuleWithLifeCycle,
	)

	require.NoError(t, app.Start(ctx))
	defer app.Stop(ctx)

	_, err := db.Exec(`INSERT INTO messages (phone_number, content) 
				VALUES('+1234567890','hello'),('+9876543210','world')`)
	if err != nil {
		t.Fatal(err)
	}

	require.Eventually(t, func() bool {
		var c int
		err := db.QueryRow(`SELECT count(*) FROM messages WHERE status != 'sent'`).Scan(&c)
		if err != nil {
			// optionally log or handle the error
			return false
		}
		return c == 0
	}, 10*time.Second, 100*time.Millisecond, "all messages should be sent")
}
