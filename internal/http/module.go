package http

import (
	"context"
	"net/http"

	"go.uber.org/fx"
)

func StartHTTP(lc fx.Lifecycle, router http.Handler) {
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go server.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}

var Module = fx.Invoke(StartHTTP)
