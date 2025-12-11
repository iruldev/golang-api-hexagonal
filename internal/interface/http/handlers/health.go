// Package handlers contains HTTP request handlers for the API.
package handlers

import (
	"context"
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// HealthData represents the health check data.
type HealthData struct {
	Status string `json:"status"`
}

// HealthHandler returns the health status of the service.
// Response format: {"success": true, "data": {"status": "ok"}}
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response.Success(w, HealthData{Status: "ok"})
}

// DBHealthChecker checks database health.
type DBHealthChecker interface {
	Ping(ctx context.Context) error
}

// ReadyzHandler handles readiness probe requests.
type ReadyzHandler struct {
	dbChecker DBHealthChecker
}

// NewReadyzHandler creates a new ReadyzHandler with optional DB checker.
func NewReadyzHandler(dbChecker DBHealthChecker) *ReadyzHandler {
	return &ReadyzHandler{dbChecker: dbChecker}
}

// ServeHTTP handles the readiness check request.
// Returns 200 if service is ready, 503 if database is unavailable.
func (h *ReadyzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check database if available
	if h.dbChecker != nil {
		if err := h.dbChecker.Ping(ctx); err != nil {
			response.Error(w, http.StatusServiceUnavailable, response.ErrServiceUnavailable, "database unavailable")
			return
		}
	}

	response.Success(w, HealthData{Status: "ready"})
}
