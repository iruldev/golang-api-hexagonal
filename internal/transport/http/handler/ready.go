package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// DatabaseChecker provides database health check capability.
type DatabaseChecker interface {
	Ping(ctx context.Context) error
}

// ReadyResponse represents the response for the ready endpoint.
type ReadyResponse struct {
	Data ReadyData `json:"data"`
}

// ReadyData contains the readiness status and checks.
type ReadyData struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// ReadyHandler handles the /ready endpoint (readiness probe).
// It checks database connectivity to determine if the service is ready.
type ReadyHandler struct {
	db DatabaseChecker
}

// NewReadyHandler creates a new ready handler.
func NewReadyHandler(db DatabaseChecker) *ReadyHandler {
	return &ReadyHandler{db: db}
}

// ServeHTTP handles the readiness check request.
func (h *ReadyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allOK := true

	// Check database connectivity
	if err := h.db.Ping(ctx); err != nil {
		checks["database"] = "failed"
		allOK = false
	} else {
		checks["database"] = "ok"
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allOK {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	resp := ReadyResponse{
		Data: ReadyData{
			Status: status,
			Checks: checks,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)
}
