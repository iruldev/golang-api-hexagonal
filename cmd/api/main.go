package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/iruldev/golang-api-hexagonal/internal/app/audit"
	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/redact"
	httpTransport "github.com/iruldev/golang-api-hexagonal/internal/transport/http"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/handler"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := contract.SetProblemBaseURL(cfg.ProblemBaseURL); err != nil {
		return fmt.Errorf("failed to set PROBLEM_BASE_URL: %w", err)
	}

	// Initialize structured JSON logger with service/env attributes
	logger := observability.NewLogger(cfg)
	slog.SetDefault(logger) // Set as default for use with slog.Info(), slog.Error(), etc.

	logger.Info("service starting",
		slog.Int("port", cfg.Port),
		slog.String("log_level", cfg.LogLevel),
		slog.Bool("otel_enabled", cfg.OTELEnabled),
	)

	// Initialize OpenTelemetry tracer provider only when enabled
	var tpShutdown func(context.Context) error
	if cfg.OTELEnabled {
		tp, err := observability.InitTracer(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize tracer: %w", err)
		}
		otel.SetTracerProvider(tp)
		tpShutdown = tp.Shutdown
		logger.Info("tracing enabled")
	} else {
		logger.Info("tracing disabled; skipping tracer provider initialization")
	}

	// Prepare database connection (non-fatal if unavailable at startup)
	db := newReconnectingDB(cfg.DatabaseURL, cfg.IgnoreDBStartupError, logger)
	defer db.Close()

	const startupPingTimeout = 5 * time.Second
	ctxPing, cancelPing := context.WithTimeout(ctx, startupPingTimeout)
	if err := db.Ping(ctxPing); err != nil {
		cancelPing()
		return fmt.Errorf("database not reachable at startup: %w", err)
	}
	cancelPing()
	logger.Info("database connected")

	// Create handlers
	healthHandler := handler.NewHealthHandler()
	readyHandler := handler.NewReadyHandler(db, logger)

	// Create user-related dependencies
	userRepo := postgres.NewUserRepo()
	idGen := postgres.NewIDGenerator()

	// Create audit-related dependencies
	redactorCfg := domain.RedactorConfig{EmailMode: cfg.AuditRedactEmail}
	piiRedactor := redact.NewPIIRedactor(redactorCfg)
	auditEventRepo := postgres.NewAuditEventRepo()
	auditService := audit.NewAuditService(auditEventRepo, piiRedactor, idGen)

	// Create a database querier using the Pool() getter for safer access
	pool := db.Pool()

	// Use a pool querier (start-up verified pool is available)
	querier := postgres.NewPoolQuerier(pool.Pool())

	// Create transaction manager
	txManager := postgres.NewTxManager(pool.Pool())

	// Create use cases
	createUserUC := user.NewCreateUserUseCase(userRepo, auditService, idGen, txManager, querier)
	getUserUC := user.NewGetUserUseCase(userRepo, querier, logger)
	listUsersUC := user.NewListUsersUseCase(userRepo, querier)

	// Create user handler
	userHandler := handler.NewUserHandler(createUserUC, getUserUC, listUsersUC)

	// Initialize Prometheus metrics registry
	metricsReg, httpMetrics := observability.NewMetricsRegistry()

	// Create router with logger for request logging middleware
	jwtConfig := httpTransport.JWTConfig{
		Enabled:   cfg.JWTEnabled,
		Secret:    []byte(cfg.JWTSecret),
		Now:       nil, // Use time.Now in production
		Issuer:    cfg.JWTIssuer,
		Audience:  cfg.JWTAudience,
		ClockSkew: cfg.JWTClockSkew,
	}
	rateLimitConfig := httpTransport.RateLimitConfig{
		RequestsPerSecond: cfg.RateLimitRPS,
		TrustProxy:        cfg.TrustProxy,
	}
	publicRouter := httpTransport.NewRouter(logger, cfg.OTELEnabled, metricsReg, httpMetrics, healthHandler, readyHandler, userHandler, cfg.MaxRequestSize, jwtConfig, rateLimitConfig)

	// Create Internal Router (Story 2.5b)
	internalRouter := httpTransport.NewInternalRouter(logger, metricsReg, httpMetrics)

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
	// Use the same timeouts as public server for now to avoid config complexity.
	// If stricter internal timeouts are needed, new config vars can be added later.
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

	// Start servers in goroutines
	serverErrors := make(chan error, 2) // buffer of 2 for both servers

	go func() {
		logger.Info("public server listening", slog.String("addr", publicAddr))
		serverErrors <- publicSrv.ListenAndServe()
	}()

	go func() {
		logger.Info("internal server listening", slog.String("addr", internalAddr))
		serverErrors <- internalSrv.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			logger.Error("server error", slog.Any("err", err))
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		logger.Info("shutdown signal received", slog.Any("signal", sig))

		// Give outstanding requests time to complete
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		// Shutdown tracer provider to flush pending spans
		if tpShutdown != nil {
			if err := tpShutdown(ctx); err != nil {
				logger.Error("tracer shutdown failed", slog.Any("err", err))
			}
		}

		// Shutdown both servers concurrently
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			if err := publicSrv.Shutdown(ctx); err != nil {
				publicSrv.Close() // Force close
				logger.Error("public server graceful shutdown failed", slog.Any("err", err))
			}
		}()

		go func() {
			defer wg.Done()
			if err := internalSrv.Shutdown(ctx); err != nil {
				internalSrv.Close() // Force close
				logger.Error("internal server graceful shutdown failed", slog.Any("err", err))
			}
		}()

		wg.Wait()
	}

	logger.Info("servers stopped gracefully")
	return nil
}

// Pooler defines the interface for a database pool.
type Pooler interface {
	Ping(context.Context) error
	Close()
	Pool() *pgxpool.Pool
}

// reconnectingDB lazily establishes a database pool and retries on readiness checks.
type reconnectingDB struct {
	dsn                string
	ignoreStartupError bool
	mu                 sync.RWMutex
	pool               Pooler
	log                *slog.Logger
	poolCreator        func(context.Context, string) (Pooler, error)
}

func newReconnectingDB(dsn string, ignoreStartupError bool, log *slog.Logger) *reconnectingDB {
	return &reconnectingDB{
		dsn:                dsn,
		ignoreStartupError: ignoreStartupError,
		log:                log,
		poolCreator: func(ctx context.Context, dsn string) (Pooler, error) {
			return postgres.NewPool(ctx, dsn)
		},
	}
}

// Ping ensures a pool exists and is healthy; recreates the pool on failure.
func (r *reconnectingDB) Ping(ctx context.Context) error {
	// Fast path: try existing pool under read lock
	r.mu.RLock()
	pool := r.pool
	r.mu.RUnlock()

	if pool == nil {
		// Create pool under write lock (double-check pattern)
		r.mu.Lock()
		if r.pool == nil {
			newPool, err := r.poolCreator(ctx, r.dsn)
			if err != nil {
				// If creation fails and we ignore startup errors
				if r.ignoreStartupError {
					r.log.Warn("database pool creation failed but IGNORE_DB_STARTUP_ERROR is set; using no-op pool", slog.Any("err", err))
					r.pool = &noopPool{}
					r.mu.Unlock()
					return nil
				}
				r.mu.Unlock()
				return err
			}
			r.pool = newPool
		}
		pool = r.pool
		r.mu.Unlock()
	}

	if err := pool.Ping(ctx); err != nil {
		if r.ignoreStartupError {
			r.log.Warn("database ping failed but IGNORE_DB_STARTUP_ERROR is set; using no-op pool", slog.Any("err", err))
			// Assign a no-op pool so checking db.Pool() doesn't panic
			// We need write lock to update r.pool
			r.mu.Lock()
			r.pool = &noopPool{}
			r.mu.Unlock()
			return nil
		}
		// pgxpool handles reconnection automatically - don't close the pool
		// Closing invalidates references held by querier/txManager causing panics
		r.log.Warn("database ping failed", slog.Any("err", err))
		return err
	}

	return nil
}

// Close shuts down the pool if it was created.
func (r *reconnectingDB) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.pool != nil {
		r.pool.Close()
		r.pool = nil
	}
}

// Pool returns the current pool for database operations.
func (r *reconnectingDB) Pool() Pooler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pool
}

// noopPool is a placeholder for when DB connection is ignored.
type noopPool struct{}

func (n *noopPool) Ping(context.Context) error { return nil }
func (n *noopPool) Close()                     {}
func (n *noopPool) Pool() *pgxpool.Pool        { return nil }
