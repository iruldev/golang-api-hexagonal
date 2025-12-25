package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

const (
	// StatusResponse constants
	StatusReady    = "ready"
	StatusNotReady = "not_ready"

	// Check names and statuses
	CheckDatabase     = "database"
	CheckStatusOk     = "ok"
	CheckStatusFailed = "failed"

	// DefaultReadyTimeout is the timeout for the readiness check
	DefaultReadyTimeout = 5 * time.Second
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
	db     DatabaseChecker
	logger *slog.Logger
}

// NewReadyHandler creates a new ready handler.
func NewReadyHandler(db DatabaseChecker, logger *slog.Logger) *ReadyHandler {
	return &ReadyHandler{
		db:     db,
		logger: logger,
	}
}

// ServeHTTP handles the readiness check request.
func (h *ReadyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), DefaultReadyTimeout)
	defer cancel()

	checks := make(map[string]string)
	allOK := true

	// Check database connectivity
	if err := h.db.Ping(ctx); err != nil {
		checks[CheckDatabase] = CheckStatusFailed
		allOK = false
	} else {
		checks[CheckDatabase] = CheckStatusOk
	}

	status := StatusReady
	httpStatus := http.StatusOK
	if !allOK {
		status = StatusNotReady
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to write ready response", slog.Any("error", err))
	}
}
