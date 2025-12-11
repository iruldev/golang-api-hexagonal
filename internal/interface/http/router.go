// Package http provides HTTP server and routing functionality.
package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/handlers"
)

// NewRouter creates a new chi router with versioned API routes.
// The router mounts all API endpoints under /api/v1 prefix for versioning.
//
// The cfg parameter is passed for future middleware configuration:
// - Story 3.3: Logging middleware (cfg.Log.Level, cfg.Log.Format)
// - Story 3.5: OpenTelemetry middleware (cfg.Observability)
// Currently unused but kept for upcoming middleware stories.
func NewRouter(cfg *config.Config) chi.Router {
	// TODO: Wire cfg into middleware chain (Stories 3.2-3.5)
	_ = cfg // Silence unused variable warning until middleware is added

	r := chi.NewRouter()

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handlers.HealthHandler)
	})

	return r
}
