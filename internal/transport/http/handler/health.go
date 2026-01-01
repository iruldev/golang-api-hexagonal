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
