package api

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/lazerion/outbox-relayer/internal/api/handler"
	"github.com/lazerion/outbox-relayer/internal/schedule"
	"github.com/lazerion/outbox-relayer/internal/service"
)

var Module = fx.Module(
	"api",

	fx.Provide(
		func(s service.QueryServiceInterface) *handler.QueryHandler {
			return handler.NewQueryHandler(s)
		},
		func(s schedule.SchedulerInterface) *handler.SchedulerHandler {
			return handler.NewSchedulerHandler(s)
		},
	),

	fx.Provide(
		func(
			schedHandler *handler.SchedulerHandler,
			queryHandler *handler.QueryHandler,
		) http.Handler {
			return NewRouter(schedHandler, queryHandler)
		},
	),
)
