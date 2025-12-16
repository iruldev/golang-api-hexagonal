// Package http provides HTTP transport layer components.
package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/iruldev/golang-api-hexagonal/internal/shared/metrics"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

// NewRouter creates a new chi router with the provided handlers and logger.
func NewRouter(logger *slog.Logger, tracingEnabled bool, metricsReg *prometheus.Registry, httpMetrics metrics.HTTPMetrics, healthHandler, readyHandler http.Handler) chi.Router {
	r := chi.NewRouter()

	// Middleware stack (order matters!):
	// 1. RequestID: Generate/passthrough request ID FIRST
	// 2. Tracing: Create spans and propagate trace context
	// 3. Metrics: Record request counts and durations (needs route pattern)
	// 4. Logging: Needs requestId and traceId in context
	// 5. RealIP: Extract real IP from headers
	// 6. Recoverer: Panic recovery
	r.Use(middleware.RequestID)
	if tracingEnabled {
		r.Use(middleware.Tracing)
	}
	r.Use(middleware.Metrics(httpMetrics))
	r.Use(middleware.RequestLogger(logger))
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)

	// Metrics endpoint (no auth required)
	r.Handle("/metrics", promhttp.HandlerFor(metricsReg, promhttp.HandlerOpts{}))

	// Health check endpoints
	r.Get("/health", healthHandler.ServeHTTP)
	r.Get("/ready", readyHandler.ServeHTTP)

	return r
}
