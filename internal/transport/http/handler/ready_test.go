package handler

import (
	"context"
	"encoding/json"
	"errors"
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
	handler := NewReadyHandler(db)

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

	assert.Equal(t, "ready", resp.Data.Status)
	assert.Equal(t, "ok", resp.Data.Checks["database"])
}

func TestReadyHandler_DatabaseDisconnected(t *testing.T) {
	db := &mockDatabase{pingErr: errors.New("connection refused")}
	handler := NewReadyHandler(db)

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

	assert.Equal(t, "not_ready", resp.Data.Status)
	assert.Equal(t, "failed", resp.Data.Checks["database"])
}
