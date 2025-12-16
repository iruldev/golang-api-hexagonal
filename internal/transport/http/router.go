// Package http provides HTTP transport layer components.
package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates a new chi router with the provided handlers.
func NewRouter(healthHandler, readyHandler http.Handler) chi.Router {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// Health check endpoints (no middleware)
	r.Get("/health", healthHandler.ServeHTTP)
	r.Get("/ready", readyHandler.ServeHTTP)

	return r
}
