package service

import (
	"go.uber.org/fx"

	"github.com/lazerion/outbox-relayer/internal/config"
	"github.com/lazerion/outbox-relayer/internal/gateway"
	"github.com/lazerion/outbox-relayer/internal/repository"
	"github.com/lazerion/outbox-relayer/internal/schedule"
)

func NewRelayerServiceProvider(
	repo repository.MessageRepository,
	sender gateway.Sender,
	cfg *config.Config,
) schedule.Job {
	return NewRelayerService(
		repo,
		sender,
		cfg.Relayer.Batch,
		cfg.Relayer.Timeout,
		cfg.Relayer.MaxAttempts,
	)
}

func NewQueryServiceProvider(
	repo repository.QueryRepository,
) QueryServiceInterface {
	return NewQueryService(repo)
}

var Module = fx.Module(
	"service",
	fx.Provide(
		NewRelayerServiceProvider,
		NewQueryServiceProvider,
	),
)
