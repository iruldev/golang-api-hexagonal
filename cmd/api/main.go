package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
	fxmodule "github.com/iruldev/golang-api-hexagonal/internal/infra/fx"
)

func main() {
	app := fx.New(
		fxmodule.Module,
		fx.Invoke(startServers),
	)

	app.Run()
}

func startServers(
	lc fx.Lifecycle,
	publicRouter chi.Router,
	internalRouter *chi.Mux,
	cfg *config.Config,
	logger *slog.Logger,
) {
	// Create Public HTTP server
	publicAddr := fmt.Sprintf(":%d", cfg.Port)
	publicSrv := &http.Server{
		Addr:              publicAddr,
		Handler:           publicRouter,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		MaxHeaderBytes:    cfg.HTTPMaxHeaderBytes,
	}

	// Create Internal HTTP server
	internalAddr := fmt.Sprintf("%s:%d", cfg.InternalBindAddress, cfg.InternalPort)
	internalSrv := &http.Server{
		Addr:              internalAddr,
		Handler:           internalRouter,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		MaxHeaderBytes:    cfg.HTTPMaxHeaderBytes,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("public server listening", slog.String("addr", publicAddr))
				if err := publicSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("public server failed", slog.Any("err", err))
					// We could trigger shutdown here, but Fx app is running.
					// Ideally we should signal the app to stop?
					// For now, logging is standard practice in Fx OnStart hooks unless we have shutdowner.
				}
			}()

			go func() {
				logger.Info("internal server listening", slog.String("addr", internalAddr))
				if err := internalSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("internal server failed", slog.Any("err", err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping servers")
			// Shutdown both servers
			// We can do this sequentially or parallel. Parallel is better for speed.
			// But let's keep it simple.
			if err := publicSrv.Shutdown(ctx); err != nil {
				logger.Error("public server shutdown failed", slog.Any("err", err))
				publicSrv.Close()
			}
			if err := internalSrv.Shutdown(ctx); err != nil {
				logger.Error("internal server shutdown failed", slog.Any("err", err))
				internalSrv.Close()
			}
			return nil
		},
	})
}
