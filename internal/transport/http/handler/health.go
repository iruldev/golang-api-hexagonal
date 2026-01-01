// Package handler provides HTTP request handlers.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse represents the response for the health endpoint.
type HealthResponse struct {
	Data HealthData `json:"data"`
}

// HealthData contains the health status.
type HealthData struct {
	Status string `json:"status"`
}

// HealthHandler handles the /health endpoint (liveness probe).
// It does NOT check database connectivity - it only verifies the service is running.
type HealthHandler struct{}

// NewHealthHandler creates a new health handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// ServeHTTP handles the health check request.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Data: HealthData{
			Status: "ok",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// LivenessResponse represents the minimal response for the liveness probe endpoint.
// This is a flat structure per K8s liveness probe best practices.
type LivenessResponse struct {
	Status string `json:"status"`
}

// livenessResponseBytes is the pre-marshaled JSON response to avoid runtime allocations.
var livenessResponseBytes = []byte(`{"status":"alive"}`)

// LivenessHandler handles the /healthz endpoint (K8s liveness probe).
// It returns a minimal response without any dependency checks.
// This endpoint should be lightweight and return <10ms to avoid false negatives.
//
// Thread-safe: This handler is stateless and safe for concurrent use.
type LivenessHandler struct{}

// NewLivenessHandler creates a new liveness handler.
func NewLivenessHandler() *LivenessHandler {
	return &LivenessHandler{}
}

// ServeHTTP handles the liveness probe request.
// Returns 200 OK with {"status": "alive"} when the service is running.
// No database or external dependency checks are performed.
// Uses w.Write directly to avoid json.Encoder reflection overhead and allocations.
func (h *LivenessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(livenessResponseBytes)
}

// =============================================================================
// Story 3.2: Readiness Probe with Dependency Status
// =============================================================================

// ReadinessResponse is the JSON response for the readiness probe endpoint.
// It provides overall status and per-dependency health information.
type ReadinessResponse struct {
	Status string                      `json:"status"` // "healthy" or "unhealthy"
	Checks map[string]DependencyStatus `json:"checks"`
}

// DependencyStatus represents the health status of a single dependency.
type DependencyStatus struct {
	Status    string `json:"status"`          // "healthy", "unhealthy", "degraded"
	LatencyMs int64  `json:"latency_ms"`      // Check duration in milliseconds
	Error     string `json:"error,omitempty"` // Error message if unhealthy
}

// DependencyChecker provides health check for a single dependency.
// Implementations should perform lightweight checks that complete quickly.
type DependencyChecker interface {
	// Name returns a unique identifier for this dependency (e.g., "database").
	Name() string
	// CheckHealth performs the health check and returns status, latency, and error.
	// The context should be used for timeout control.
	CheckHealth(ctx context.Context) (status string, latency time.Duration, err error)
}

// DefaultCheckTimeout is the default timeout for each dependency check.
const DefaultCheckTimeout = 2 * time.Second

// ReadinessHandler handles the /readyz endpoint (K8s readiness probe).
// It checks all registered dependencies and returns 200 if all are healthy,
// or 503 if any dependency is unhealthy.
//
// Thread-safe: This handler is safe for concurrent use.
type ReadinessHandler struct {
	checkers []DependencyChecker
	timeout  time.Duration
}

// NewReadinessHandler creates a new readiness handler with the given dependency checkers.
// If timeout is 0, DefaultCheckTimeout (2s) is used.
func NewReadinessHandler(timeout time.Duration, checkers ...DependencyChecker) *ReadinessHandler {
	if timeout == 0 {
		timeout = DefaultCheckTimeout
	}
	return &ReadinessHandler{
		checkers: checkers,
		timeout:  timeout,
	}
}

// ServeHTTP handles the readiness probe request.
// Returns 200 OK if all dependencies are healthy, 503 Service Unavailable otherwise.
// The response includes per-dependency status and latency information.
func (h *ReadinessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]DependencyStatus, len(h.checkers))
	allHealthy := true

	for _, checker := range h.checkers {
		// Create timeout context for this check
		ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
		status, latency, err := checker.CheckHealth(ctx)
		cancel()

		depStatus := DependencyStatus{
			Status:    status,
			LatencyMs: latency.Milliseconds(),
		}

		if err != nil {
			depStatus.Error = err.Error()
			allHealthy = false
		} else if status != "healthy" {
			allHealthy = false
		}

		checks[checker.Name()] = depStatus
	}

	overallStatus := "healthy"
	httpStatus := http.StatusOK
	if !allHealthy {
		overallStatus = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	resp := ReadinessResponse{
		Status: overallStatus,
		Checks: checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)
}
