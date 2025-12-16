package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	httpTransport "github.com/iruldev/golang-api-hexagonal/internal/transport/http"
	"github.com/stretchr/testify/assert"
)

// fakeDB implements DatabaseChecker for integration-style router tests.
type fakeDB struct {
	pingErr error
}

func (f *fakeDB) Ping(ctx context.Context) error {
	return f.pingErr
}

// testLogger returns a discarding logger for use in tests.
func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

// TestIntegrationRoutes covers /health and /ready through the router with DB ok/fail.
func TestIntegrationRoutes(t *testing.T) {
	healthHandler := NewHealthHandler()
	logger := testLogger()

	t.Run("ready OK", func(t *testing.T) {
		db := &fakeDB{pingErr: nil}
		readyHandler := NewReadyHandler(db)
		r := httpTransport.NewRouter(logger, healthHandler, readyHandler)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"ready","checks":{"database":"ok"}}}`, rec.Body.String())
	})

	t.Run("ready not ready", func(t *testing.T) {
		db := &fakeDB{pingErr: assert.AnError}
		readyHandler := NewReadyHandler(db)
		r := httpTransport.NewRouter(logger, healthHandler, readyHandler)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"not_ready","checks":{"database":"failed"}}}`, rec.Body.String())
	})

	t.Run("health ok", func(t *testing.T) {
		db := &fakeDB{pingErr: nil}
		readyHandler := NewReadyHandler(db)
		r := httpTransport.NewRouter(logger, healthHandler, readyHandler)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"ok"}}`, rec.Body.String())
	})
}
