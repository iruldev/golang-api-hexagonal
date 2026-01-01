// Package handler provides HTTP request handlers.
package handler

import (
	"encoding/json"
	"net/http"
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
