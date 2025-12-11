// Package http provides HTTP server and routing functionality.
package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/handlers"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
)

// NewRouter creates a new chi router with versioned API routes.
// The router mounts all API endpoints under /api/v1 prefix for versioning.
//
// The cfg parameter is passed for future middleware configuration:
// - Story 3.3: Logging middleware (cfg.Log.Level, cfg.Log.Format)
// - Story 3.5: OpenTelemetry middleware (cfg.Observability)
func NewRouter(cfg *config.Config) chi.Router {
	// TODO: Wire cfg into logging/otel middleware (Stories 3.3, 3.5)
	_ = cfg

	r := chi.NewRouter()

	// Global middleware (Story 3.2)
	r.Use(middleware.RequestID)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handlers.HealthHandler)
	})

	return r
}
