package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

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
