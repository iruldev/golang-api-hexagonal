// Package http provides HTTP transport layer components.
package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

// NewRouter creates a new chi router with the provided handlers and logger.
func NewRouter(logger *slog.Logger, healthHandler, readyHandler http.Handler) chi.Router {
	r := chi.NewRouter()

	// Middleware stack (order matters!)
	// 1. RequestID: Generate/passthrough request ID FIRST
	// 2. Logging: Needs requestId in context
	// 3. RealIP: Extract real IP from headers
	// 4. Recoverer: Panic recovery
	r.Use(middleware.RequestID)
	r.Use(middleware.RequestLogger(logger))
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)

	// Health check endpoints
	r.Get("/health", healthHandler.ServeHTTP)
	r.Get("/ready", readyHandler.ServeHTTP)

	return r
}
