// Package fxmodule provides Uber Fx dependency injection modules for the application.
// This replaces Google Wire which is incompatible with pgx/puddle packages.
//
// Usage in main.go:
//
//	app := fx.New(
//	    fxmodule.Module,
//	    fx.Invoke(run),
//	)
//	app.Run()
package fxmodule

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-chi/chi/v5"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"

	"github.com/iruldev/golang-api-hexagonal/internal/app/audit"
	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/resilience"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/metrics"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/redact"
	httpTransport "github.com/iruldev/golang-api-hexagonal/internal/transport/http"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/handler"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

// Module provides all application dependencies via Uber Fx.
var Module = fx.Options(
	ConfigModule,
	ObservabilityModule,
	ResilienceModule,
	PostgresModule,
	DomainModule,
	AppModule,
	TransportModule,
)

// ConfigModule provides configuration dependencies.
var ConfigModule = fx.Options(
	fx.Provide(config.Load),
	fx.Invoke(func(cfg *config.Config) error {
		return contract.SetProblemBaseURL(cfg.ProblemBaseURL)
	}),
)

// ObservabilityModule provides logging and metrics dependencies.
var ObservabilityModule = fx.Options(
	fx.Provide(observability.NewLogger),
	fx.Invoke(func(logger *slog.Logger) {
		slog.SetDefault(logger)
	}),
	fx.Provide(provideMetrics),
	fx.Provide(provideTracer),
)

func provideTracer(lc fx.Lifecycle, cfg *config.Config, logger *slog.Logger) (*sdktrace.TracerProvider, error) {
	if !cfg.OTELEnabled {
		logger.Info("tracing disabled")
		return sdktrace.NewTracerProvider(), nil
	}

	tp, err := observability.InitTracer(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)
	logger.Info("tracing enabled")

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("shutting down tracer")
			return tp.Shutdown(ctx)
		},
	})

	return tp, nil
}

// MetricsResult holds Prometheus metrics components.
type MetricsResult struct {
	fx.Out
	Registry    *prometheus.Registry
	HTTPMetrics metrics.HTTPMetrics
}

func provideMetrics() MetricsResult {
	reg, httpMetrics := observability.NewMetricsRegistry()
	return MetricsResult{
		Registry:    reg,
		HTTPMetrics: httpMetrics,
	}
}

// ResilienceModule provides resilience dependencies (Story 1.6 + 1.7).
var ResilienceModule = fx.Options(
	fx.Provide(provideResilienceConfig),
	// Circuit Breaker components
	fx.Provide(provideCircuitBreakerMetrics),
	fx.Provide(provideCircuitBreakerPresets),
	// Retry components
	fx.Provide(provideRetryMetrics),
	fx.Provide(provideRetrier),
	// Timeout components
	fx.Provide(provideTimeoutMetrics),
	fx.Provide(provideTimeoutPresets),
	// Bulkhead components
	fx.Provide(provideBulkheadMetrics),
	fx.Provide(provideBulkheadPresets),
	// Shutdown components
	fx.Provide(provideShutdownMetrics),
	fx.Provide(provideShutdownCoordinator),
	// ResilienceWrapper (composes all patterns)
	fx.Provide(provideResilienceWrapper),
)

func provideResilienceConfig(cfg *config.Config) resilience.ResilienceConfig {
	return resilience.NewResilienceConfig(cfg)
}

func provideCircuitBreakerMetrics(registry *prometheus.Registry) *resilience.CircuitBreakerMetrics {
	return resilience.NewCircuitBreakerMetrics(registry)
}

func provideCircuitBreakerPresets(
	resCfg resilience.ResilienceConfig,
	metrics *resilience.CircuitBreakerMetrics,
	logger *slog.Logger,
) *resilience.CircuitBreakerPresets {
	return resilience.NewCircuitBreakerPresets(
		resCfg.CircuitBreaker,
		resilience.WithMetrics(metrics),
		resilience.WithLogger(logger),
	)
}

func provideRetryMetrics(registry *prometheus.Registry) *resilience.RetryMetrics {
	return resilience.NewRetryMetrics(registry)
}

func provideRetrier(
	resCfg resilience.ResilienceConfig,
	metrics *resilience.RetryMetrics,
	logger *slog.Logger,
) resilience.Retrier {
	return resilience.NewRetrier(
		"default",
		resCfg.Retry,
		resilience.WithRetryMetrics(metrics),
		resilience.WithRetryLogger(logger),
	)
}

func provideTimeoutMetrics(registry *prometheus.Registry) *resilience.TimeoutMetrics {
	return resilience.NewTimeoutMetrics(registry)
}

func provideTimeoutPresets(
	resCfg resilience.ResilienceConfig,
	metrics *resilience.TimeoutMetrics,
	logger *slog.Logger,
) *resilience.TimeoutPresets {
	return resilience.NewTimeoutPresets(
		resCfg.Timeout,
		resilience.WithTimeoutMetrics(metrics),
		resilience.WithTimeoutLogger(logger),
	)
}

func provideBulkheadMetrics(registry *prometheus.Registry) *resilience.BulkheadMetrics {
	return resilience.NewBulkheadMetrics(registry)
}

func provideBulkheadPresets(
	resCfg resilience.ResilienceConfig,
	metrics *resilience.BulkheadMetrics,
	logger *slog.Logger,
) *resilience.BulkheadPresets {
	return resilience.NewBulkheadPresets(
		resCfg.Bulkhead,
		resilience.WithBulkheadMetrics(metrics),
		resilience.WithBulkheadLogger(logger),
	)
}

func provideShutdownMetrics(registry *prometheus.Registry) *resilience.ShutdownMetrics {
	return resilience.NewShutdownMetrics(registry)
}

func provideShutdownCoordinator(
	resCfg resilience.ResilienceConfig,
	metrics *resilience.ShutdownMetrics,
	logger *slog.Logger,
) resilience.ShutdownCoordinator {
	return resilience.NewShutdownCoordinator(
		resCfg.Shutdown,
		resilience.WithShutdownMetrics(metrics),
		resilience.WithShutdownLogger(logger),
	)
}

func provideResilienceWrapper(
	cbPresets *resilience.CircuitBreakerPresets,
	retrier resilience.Retrier,
	timeoutPresets *resilience.TimeoutPresets,
	bulkheadPresets *resilience.BulkheadPresets,
	logger *slog.Logger,
) resilience.ResilienceWrapper {
	return resilience.NewResilienceWrapper(
		resilience.WithCircuitBreakerFactory(cbPresets.Factory()),
		resilience.WithWrapperRetrier(retrier),
		resilience.WithWrapperTimeout(timeoutPresets.Default()),
		resilience.WithWrapperBulkhead(bulkheadPresets.Default()),
		resilience.WithWrapperLogger(logger),
	)
}

// PostgresModule provides database dependencies.
var PostgresModule = fx.Options(
	fx.Provide(providePoolConfig),
	fx.Provide(providePool),
	fx.Provide(provideQuerier),
	fx.Provide(provideTxManager),
	// Idempotency storage (Story 2.5)
	fx.Provide(provideIdempotencyRepo),
	fx.Provide(provideIdempotencyStore),
	fx.Invoke(startIdempotencyCleaner),
)

func providePoolConfig(cfg *config.Config) postgres.PoolConfig {
	return postgres.PoolConfig{
		MaxConns:        cfg.DBPoolMaxConns,
		MinConns:        cfg.DBPoolMinConns,
		MaxConnLifetime: cfg.DBPoolMaxLifetime,
	}
}

func providePool(lc fx.Lifecycle, cfg *config.Config, poolCfg postgres.PoolConfig, logger *slog.Logger) (postgres.Pooler, error) {
	ctx := context.Background()
	pool := postgres.NewResilientPool(ctx, cfg.DatabaseURL, poolCfg, cfg.IgnoreDBStartupError, logger)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing database pool")
			pool.Close()
			return nil
		},
	})

	return pool, nil
}

func provideQuerier(pool postgres.Pooler) domain.Querier {
	if pool == nil {
		return nil
	}
	return postgres.NewPoolQuerier(pool)
}

func provideTxManager(pool postgres.Pooler) domain.TxManager {
	if pool == nil {
		return nil
	}
	return postgres.NewTxManager(pool)
}

func provideIdempotencyRepo(pool postgres.Pooler) *postgres.IdempotencyRepo {
	return postgres.NewIdempotencyRepo(pool)
}

func provideIdempotencyStore(repo *postgres.IdempotencyRepo) middleware.IdempotencyStore {
	return repo
}

func startIdempotencyCleaner(
	lc fx.Lifecycle,
	pool postgres.Pooler,
	cfg *config.Config,
	logger *slog.Logger,
	registry *prometheus.Registry,
) {
	cleaner := postgres.NewIdempotencyCleaner(
		pool,
		postgres.IdempotencyCleanerConfig{
			Interval: cfg.IdempotencyCleanupInterval,
		},
		logger,
		registry,
	)

	lc.Append(fx.Hook{
		OnStart: cleaner.Start,
		OnStop:  cleaner.Stop,
	})
}

// DomainModule provides domain-level dependencies.
var DomainModule = fx.Options(
	fx.Provide(
		fx.Annotate(
			postgres.NewUserRepo,
			fx.As(new(domain.UserRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			postgres.NewAuditEventRepo,
			fx.As(new(domain.AuditEventRepository)),
		),
	),
	fx.Provide(postgres.NewIDGenerator),
	fx.Provide(provideRedactorConfig),
	fx.Provide(
		fx.Annotate(
			redact.NewPIIRedactor,
			fx.As(new(domain.Redactor)),
		),
	),
)

func provideRedactorConfig(cfg *config.Config) domain.RedactorConfig {
	return domain.RedactorConfig{EmailMode: cfg.AuditRedactEmail}
}

// AppModule provides application layer dependencies.
var AppModule = fx.Options(
	fx.Provide(audit.NewAuditService),
	fx.Provide(user.NewCreateUserUseCase),
	fx.Provide(user.NewGetUserUseCase),
	fx.Provide(user.NewListUsersUseCase),
)

// TransportModule provides HTTP transport dependencies.
var TransportModule = fx.Options(
	fx.Provide(handler.NewHealthHandler),
	fx.Provide(provideReadyHandler),
	fx.Provide(provideStartupHandler),
	fx.Provide(provideUserHandler),
	fx.Provide(provideJWTConfig),
	fx.Provide(provideRateLimitConfig),
	// Story 3.4: Health Check Library Integration
	fx.Provide(provideHealthCheckRegistry),
	fx.Provide(providePublicRouter),
	fx.Provide(provideInternalRouter),
	fx.Invoke(registerStartupHook),
)

func provideReadyHandler(pool postgres.Pooler, logger *slog.Logger) *handler.ReadyHandler {
	return handler.NewReadyHandler(pool, logger)
}

// provideHealthCheckRegistry creates the HealthCheckRegistry with standard checks.
// Story 3.4: Health Check Library Integration.
func provideHealthCheckRegistry(
	cfg *config.Config,
	registry *prometheus.Registry,
	pool postgres.Pooler,
) *handler.HealthCheckRegistry {
	hc := handler.NewHealthCheckRegistry(registry, "app")

	// Add standard liveness check - goroutine count threshold
	// Detects goroutine leaks that could indicate unhealthy state
	hc.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(10000))

	// Add database readiness check with configurable timeout
	// Uses the library-compatible check function
	hc.AddReadinessCheck("database", postgres.NewDatabaseCheck(pool, cfg.HealthCheckDBTimeout))

	return hc
}

func provideUserHandler(
	createUC *user.CreateUserUseCase,
	getUC *user.GetUserUseCase,
	listUC *user.ListUsersUseCase,
) *handler.UserHandler {
	return handler.NewUserHandler(createUC, getUC, listUC, httpTransport.BasePath+"/users")
}

func provideJWTConfig(cfg *config.Config) httpTransport.JWTConfig {
	return httpTransport.JWTConfig{
		Enabled:   cfg.JWTEnabled,
		Secret:    []byte(cfg.JWTSecret),
		Now:       nil,
		Issuer:    cfg.JWTIssuer,
		Audience:  cfg.JWTAudience,
		ClockSkew: cfg.JWTClockSkew,
	}
}

func provideRateLimitConfig(cfg *config.Config) httpTransport.RateLimitConfig {
	return httpTransport.RateLimitConfig{
		RequestsPerSecond: cfg.RateLimitRPS,
		TrustProxy:        cfg.TrustProxy,
	}
}

func providePublicRouter(
	cfg *config.Config,
	logger *slog.Logger,
	registry *prometheus.Registry,
	httpMetrics metrics.HTTPMetrics,
	healthRegistry *handler.HealthCheckRegistry, // Story 3.4: Use library registry
	healthHandler *handler.HealthHandler,
	readyHandler *handler.ReadyHandler,
	startupHandler *handler.StartupHandler,
	userHandler *handler.UserHandler,
	jwtConfig httpTransport.JWTConfig,
	rateLimitConfig httpTransport.RateLimitConfig,
	shutdownCoord resilience.ShutdownCoordinator,
	idempotencyStore middleware.IdempotencyStore,
) chi.Router {
	return httpTransport.NewRouter(
		logger,
		cfg.OTELEnabled,
		registry,
		httpMetrics,
		httpTransport.RouterHandlers{
			LivenessHandler:  healthRegistry.LiveHandler(), // Story 3.4: Library handler
			HealthHandler:    healthHandler,
			ReadyHandler:     readyHandler,
			ReadinessHandler: healthRegistry.ReadyHandler(), // Story 3.4: Library handler
			StartupHandler:   startupHandler,
			UserHandler:      userHandler,
		},
		cfg.MaxRequestSize,
		jwtConfig,
		rateLimitConfig,
		shutdownCoord,
		idempotencyStore,
		cfg.IdempotencyTTL,
	)
}

func provideInternalRouter(
	logger *slog.Logger,
	registry *prometheus.Registry,
	httpMetrics metrics.HTTPMetrics,
) *chi.Mux {
	return httpTransport.NewInternalRouter(logger, registry, httpMetrics)
}

func provideStartupHandler() *handler.StartupHandler {
	return handler.NewStartupHandler()
}

func registerStartupHook(
	lc fx.Lifecycle,
	h *handler.StartupHandler,
	logger *slog.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Story 3.3: Mark startup probe as ready when application starts.
			// This signals to K8s that the application initialization is complete.
			h.MarkReady()
			logger.Info("startup probe marked ready", slog.Bool("ready", true))
			return nil
		},
	})
}
