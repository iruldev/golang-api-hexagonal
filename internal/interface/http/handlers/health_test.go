package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockDBChecker is a mock implementation of DBHealthChecker
type mockDBChecker struct {
	err error
}

func (m *mockDBChecker) Ping(ctx context.Context) error {
	return m.err
}

func TestReadyzHandler_HealthyDB(t *testing.T) {
	handler := NewReadyzHandler(&mockDBChecker{err: nil})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ready")
}

func TestReadyzHandler_UnhealthyDB(t *testing.T) {
	handler := NewReadyzHandler(&mockDBChecker{err: errors.New("connection refused")})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "database unavailable")
}

func TestReadyzHandler_NilDBChecker(t *testing.T) {
	handler := NewReadyzHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// With nil DB checker, should still return 200 (DB is optional)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ready")
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ok")
}
