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

	// Middleware stack
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestLogger(logger)) // Structured JSON request logging

	// Health check endpoints
	r.Get("/health", healthHandler.ServeHTTP)
	r.Get("/ready", readyHandler.ServeHTTP)

	return r
}
