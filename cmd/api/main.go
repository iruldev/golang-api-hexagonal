package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
	fxmodule "github.com/iruldev/golang-api-hexagonal/internal/infra/fx"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/resilience"
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
	coord resilience.ShutdownCoordinator,
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
			start := time.Now()
			logger.Info("initiating graceful shutdown",
				slog.Duration("drain_period", coord.Config().DrainPeriod),
				slog.Int64("active_requests", coord.ActiveCount()),
			)

			// Phase 1: Stop accepting new requests
			coord.InitiateShutdown()

			// Phase 2: Wait for in-flight requests to complete
			drainTimeout := false
			if err := coord.WaitForDrain(ctx); err != nil {
				logger.Warn("drain incomplete, proceeding with shutdown",
					slog.Any("error", err),
					slog.Int64("remaining_requests", coord.ActiveCount()),
				)
				drainTimeout = true
			}

			// Phase 3: Shutdown HTTP servers
			// Use GracePeriod for additional cleanup after drain
			serverCtx, cancel := context.WithTimeout(ctx, coord.Config().GracePeriod)
			defer cancel()

			if err := publicSrv.Shutdown(serverCtx); err != nil {
				logger.Error("public server shutdown failed", slog.Any("err", err))
				_ = publicSrv.Close()
			}
			if err := internalSrv.Shutdown(serverCtx); err != nil {
				logger.Error("internal server shutdown failed", slog.Any("err", err))
				_ = internalSrv.Close()
			}

			duration := time.Since(start)
			if drainTimeout {
				logger.Warn("graceful shutdown completed with timeout",
					slog.Duration("duration", duration),
					slog.String("status", "timeout"),
				)
				// Return error to signal timeout to Fx. Fx will exit with non-zero code
				// when OnStop returns an error. This is cleaner than os.Exit(1) which
				// would bypass remaining cleanup hooks.
				return fmt.Errorf("graceful shutdown timed out after %v", duration)
			}

			logger.Info("graceful shutdown completed successfully",
				slog.Duration("duration", duration),
				slog.String("status", "success"),
			)

			return nil
		},
	})
}
