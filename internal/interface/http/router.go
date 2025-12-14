// Package http provides HTTP server and routing functionality.
package http

import (
	"context"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/handlers"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// TracerShutdown holds the tracer shutdown function for graceful cleanup.
var TracerShutdown func(context.Context) error

// RouterDeps holds dependencies for the router.
type RouterDeps struct {
	Config          *config.Config
	DBChecker       handlers.DBHealthChecker // Optional, can be nil
	RedisChecker    handlers.DBHealthChecker // Optional, can be nil
	KafkaChecker    handlers.DBHealthChecker // Optional, can be nil (Story 13.1)
	RabbitMQChecker handlers.DBHealthChecker // Optional, can be nil (Story 13.2)
	Authenticator   middleware.Authenticator // Optional, can be nil (Story 14.1)
	// AdminFeatureFlagProvider for feature flag management API (Story 15.2)
	AdminFeatureFlagProvider runtimeutil.AdminFeatureFlagProvider
	// UserRoleProvider for user role management API (Story 15.3)
	UserRoleProvider runtimeutil.UserRoleProvider
	// QueueInspector for job queue inspection API (Story 15.4)
	QueueInspector runtimeutil.QueueInspector
}

// NewRouter creates a new chi router with versioned API routes.
// The router mounts all API endpoints under /api/v1 prefix for versioning.
//
// The deps parameter provides configuration and dependencies:
// - Config drives middleware configuration
// - DBChecker is used for the /readyz endpoint (can be nil)
//
// Route Registration:
// All routes are registered via RegisterRoutes() in routes.go (Story 3.6)
// See routes.go for documentation on adding new handlers.
func NewRouter(deps RouterDeps) chi.Router {
	cfg := deps.Config

	// Initialize logger with config (Story 3.3)
	logger, err := observability.NewLogger(&cfg.Log, cfg.App.Env)
	if err != nil {
		log.Printf("Failed to initialize logger, using nop: %v", err)
		logger = observability.NewNopLogger()
	}

	// Initialize tracer if configured (Story 3.5)
	if cfg.Observability.ExporterEndpoint != "" {
		_, shutdown, err := observability.NewTracerProvider(context.Background(), &cfg.Observability)
		if err != nil {
			log.Printf("Failed to initialize tracer: %v", err)
		} else {
			TracerShutdown = shutdown
		}
	}

	r := chi.NewRouter()

	// Global middleware (order matters!)
	r.Use(middleware.Recovery(logger)) // Story 3.4 - FIRST to catch all panics
	r.Use(middleware.RequestID)        // Story 3.2
	r.Use(middleware.SecurityHeaders)  // Story 14.3 - Security Headers
	r.Use(middleware.Metrics)          // Story 5.5 - HTTP metrics
	r.Use(middleware.Otel("api"))      // Story 3.5 - OTEL tracing
	r.Use(middleware.Logging(logger))  // Story 3.3

	// Kubernetes health check endpoints at root level (Story 4.7)
	r.Get("/healthz", handlers.HealthHandler)
	readyzHandler := handlers.NewReadyzHandler(deps.DBChecker)
	if deps.RedisChecker != nil {
		readyzHandler = readyzHandler.WithRedis(deps.RedisChecker)
	}
	if deps.KafkaChecker != nil {
		readyzHandler = readyzHandler.WithKafka(deps.KafkaChecker)
	}
	if deps.RabbitMQChecker != nil {
		readyzHandler = readyzHandler.WithRabbitMQ(deps.RabbitMQChecker)
	}
	r.Handle("/readyz", readyzHandler)

	// Prometheus metrics endpoint (Story 5.4)
	r.Handle("/metrics", promhttp.Handler())

	// API v1 routes - delegate to routes.go (Story 3.6)
	r.Route("/api/v1", func(r chi.Router) {
		// Apply authentication middleware if authenticator is provided (Story 14.1)
		if deps.Authenticator != nil {
			r.Use(middleware.AuthMiddleware(deps.Authenticator))
		}
		RegisterRoutes(r)
	})

	// Admin routes - separated from versioned API (Story 15.1)
	// Requires authentication + admin role to access any endpoint
	if deps.Authenticator != nil {
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(deps.Authenticator))
			r.Use(middleware.RequireRole("admin"))
			adminDeps := AdminDeps{
				FeatureFlagProvider: deps.AdminFeatureFlagProvider,
				UserRoleProvider:    deps.UserRoleProvider,
				QueueInspector:      deps.QueueInspector,
				Logger:              logger,
			}
			RegisterAdminRoutes(r, adminDeps)
		})
	}

	return r
}
