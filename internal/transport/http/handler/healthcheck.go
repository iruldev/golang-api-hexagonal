package handler

import (
	"net/http"

	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
)

// =============================================================================
// Story 3.4: Health Check Library Integration
// =============================================================================

// HealthCheckRegistry wraps heptiolabs/healthcheck with project-specific patterns.
// It provides a unified interface for registering liveness and readiness checks
// using the library's parallel execution and timeout capabilities.
//
// Thread-safe: The underlying library handler is safe for concurrent use.
type HealthCheckRegistry struct {
	handler healthcheck.Handler
}

// NewHealthCheckRegistry creates a new health check registry with Prometheus metrics.
// The namespace parameter is used for Prometheus metric naming.
func NewHealthCheckRegistry(registry prometheus.Registerer, namespace string) *HealthCheckRegistry {
	return &HealthCheckRegistry{
		handler: healthcheck.NewMetricsHandler(registry, namespace),
	}
}

// NewHealthCheckRegistrySimple creates a new health check registry without metrics.
// Useful for testing or when Prometheus integration is not needed.
func NewHealthCheckRegistrySimple() *HealthCheckRegistry {
	return &HealthCheckRegistry{
		handler: healthcheck.NewHandler(),
	}
}

// AddLivenessCheck registers a liveness check.
// Liveness checks indicate that the application should be restarted.
// Every liveness check is also included as a readiness check.
func (r *HealthCheckRegistry) AddLivenessCheck(name string, check healthcheck.Check) {
	r.handler.AddLivenessCheck(name, check)
}

// AddReadinessCheck registers a readiness check.
// Readiness checks indicate that the application can serve traffic.
// A failed readiness check means the instance should not receive requests,
// but should not necessarily be restarted.
func (r *HealthCheckRegistry) AddReadinessCheck(name string, check healthcheck.Check) {
	r.handler.AddReadinessCheck(name, check)
}

// LiveHandler returns the HTTP handler for the /healthz liveness endpoint.
// The handler returns 200 OK if all liveness checks pass, 503 otherwise.
// Response format is JSON with error messages for failed checks.
func (r *HealthCheckRegistry) LiveHandler() http.HandlerFunc {
	return r.handler.LiveEndpoint
}

// ReadyHandler returns the HTTP handler for the /readyz readiness endpoint.
// The handler returns 200 OK if all readiness checks pass, 503 otherwise.
// Response format is JSON with error messages for failed checks.
func (r *HealthCheckRegistry) ReadyHandler() http.HandlerFunc {
	return r.handler.ReadyEndpoint
}

// Handler returns the underlying healthcheck.Handler for advanced usage.
// This can be used to serve both /live and /ready endpoints from a single handler.
func (r *HealthCheckRegistry) Handler() healthcheck.Handler {
	return r.handler
}
