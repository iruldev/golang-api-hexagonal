package handler

import (
	"net/http"
)

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
