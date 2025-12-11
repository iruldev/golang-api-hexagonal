// Package http provides HTTP server and routing functionality.
package http

import (
	"context"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// TracerShutdown holds the tracer shutdown function for graceful cleanup.
var TracerShutdown func(context.Context) error

// NewRouter creates a new chi router with versioned API routes.
// The router mounts all API endpoints under /api/v1 prefix for versioning.
//
// The cfg parameter drives middleware configuration:
// - Logging middleware (cfg.Log.Level, cfg.Log.Format, cfg.App.Env)
// - OpenTelemetry middleware (cfg.Observability) - Story 3.5
//
// Route Registration:
// All routes are registered via RegisterRoutes() in routes.go (Story 3.6)
// See routes.go for documentation on adding new handlers.
func NewRouter(cfg *config.Config) chi.Router {
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
	r.Use(middleware.Otel("api"))      // Story 3.5 - OTEL tracing
	r.Use(middleware.Logging(logger))  // Story 3.3

	// API v1 routes - delegate to routes.go (Story 3.6)
	r.Route("/api/v1", RegisterRoutes)

	return r
}
