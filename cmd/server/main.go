package main

import (
	"github.com/lazerion/outbox-relayer/internal/api"
	"github.com/lazerion/outbox-relayer/internal/cache"
	"github.com/lazerion/outbox-relayer/internal/gateway"
	"github.com/lazerion/outbox-relayer/internal/http"
	"github.com/lazerion/outbox-relayer/internal/infra"
	"github.com/lazerion/outbox-relayer/internal/repository"
	"github.com/lazerion/outbox-relayer/internal/schedule"
	"github.com/lazerion/outbox-relayer/internal/service"
	"go.uber.org/fx"

	"github.com/lazerion/outbox-relayer/internal/config"
)

func main() {
	fx.New(
		config.Module,
		repository.Module,
		gateway.Module,
		service.Module,
		schedule.Module,
		api.Module,
		http.Module,
		infra.Module,
		schedule.ModuleWithLifeCycle,
		cache.Module,
	).Run()
}
