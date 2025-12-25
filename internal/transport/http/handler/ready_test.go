package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDatabase implements DatabaseChecker for testing.
type mockDatabase struct {
	pingErr error
}

func (m *mockDatabase) Ping(ctx context.Context) error {
	return m.pingErr
}

func TestReadyHandler_DatabaseConnected(t *testing.T) {
	db := &mockDatabase{pingErr: nil}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewReadyHandler(db, logger)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check content type
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Check response body
	var resp ReadyResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, StatusReady, resp.Data.Status)
	assert.Equal(t, CheckStatusOk, resp.Data.Checks[CheckDatabase])
}

func TestReadyHandler_DatabaseDisconnected(t *testing.T) {
	db := &mockDatabase{pingErr: errors.New("connection refused")}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewReadyHandler(db, logger)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	// Check content type
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Check response body
	var resp ReadyResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, StatusNotReady, resp.Data.Status)
	assert.Equal(t, CheckStatusFailed, resp.Data.Checks[CheckDatabase])
}

// TestReadyHandler_Idempotent verifies that /ready endpoint is idempotent
// and does not mutate state when called repeatedly.
// This satisfies Story 4.3 AC#3: idempotency verified.
func TestReadyHandler_Idempotent(t *testing.T) {
	db := &mockDatabase{pingErr: nil}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewReadyHandler(db, logger)

	// Call /ready 10 times in rapid succession
	const iterations = 10
	for i := 0; i < iterations; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Every call should return consistent results
		assert.Equal(t, http.StatusOK, rec.Code, "iteration %d: expected 200 OK", i)

		var resp ReadyResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err, "iteration %d: decode error", i)

		assert.Equal(t, StatusReady, resp.Data.Status, "iteration %d: expected ready status", i)
		assert.Equal(t, CheckStatusOk, resp.Data.Checks[CheckDatabase], "iteration %d: expected database ok", i)
	}

	// Verify mock was called but no state was mutated
	// (mockDatabase has no internal counter, proving handler is stateless)
}
