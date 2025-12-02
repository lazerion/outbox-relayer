package api

import (
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lazerion/outbox-relayer/internal/api/docs"
	"github.com/lazerion/outbox-relayer/internal/api/handler"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(
	schedHandler *handler.SchedulerHandler,
	queryHandler *handler.QueryHandler,
) http.Handler {

	r := mux.NewRouter()

	// --------------------------------
	// API v1
	// --------------------------------
	v1 := r.PathPrefix("/api/v1").Subrouter()

	// Scheduler endpoints
	v1.HandleFunc("/scheduler/toggle", schedHandler.ToggleScheduler).
		Methods(http.MethodPost)

	// Query endpoints
	v1.HandleFunc("/messages/sent", queryHandler.ListSentMessages).
		Methods(http.MethodGet)

	// Swagger endpoint
	v1.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	return r
}
