package handler

import (
	"net/http"
	"sync/atomic"
)

// =============================================================================
// Story 3.3: Startup Probe Implementation
// =============================================================================

// Pre-marshaled JSON responses for zero allocations.
var (
	startupStartingBytes = []byte(`{"status":"starting"}`)
	startupReadyBytes    = []byte(`{"status":"ready"}`)
)

// StartupHandler handles the /startupz endpoint (K8s startup probe).
// Returns 503 Service Unavailable until MarkReady() is called, then returns 200 OK.
// This allows Kubernetes to wait for slow-starting containers without killing them.
//
// Thread-safe: Uses atomic.Bool for lock-free state transitions.
type StartupHandler struct {
	ready atomic.Bool
}

// NewStartupHandler creates a new startup handler.
// Initial state is NOT ready (returns 503 Service Unavailable).
func NewStartupHandler() *StartupHandler {
	return &StartupHandler{}
}

// MarkReady signals that startup is complete.
// After this call, ServeHTTP returns 200 OK.
// Thread-safe and idempotent.
func (h *StartupHandler) MarkReady() {
	h.ready.Store(true)
}

// IsReady returns true if startup is complete.
// Thread-safe.
func (h *StartupHandler) IsReady() bool {
	return h.ready.Load()
}

// ServeHTTP handles the startup probe request.
// Returns 503 Service Unavailable with {"status":"starting"} if still initializing,
// or 200 OK with {"status":"ready"} once MarkReady() has been called.
// Uses pre-marshaled bytes to achieve zero allocations per request.
func (h *StartupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.ready.Load() {
		w.WriteHeader(http.StatusOK)
		w.Write(startupReadyBytes)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write(startupStartingBytes)
	}
}
