package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	httpx "github.com/iruldev/golang-api-hexagonal/internal/interface/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testConfig returns a minimal config for testing.
func testConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env: "test",
		},
		Log: config.LogConfig{
			Level:  "debug",
			Format: "console",
		},
	}
}

func TestHealthEndpoint(t *testing.T) {
	router := httpx.NewRouter(testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Response is now envelope format: {success: true, data: {status: ok}}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &envelope)
	require.NoError(t, err)
	assert.True(t, envelope.Success)
	assert.Equal(t, "ok", envelope.Data.Status)
}

func TestHealthEndpoint_MethodNotAllowed(t *testing.T) {
	router := httpx.NewRouter(testConfig())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestNonExistentRoute(t *testing.T) {
	router := httpx.NewRouter(testConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAPIVersionPrefix(t *testing.T) {
	router := httpx.NewRouter(testConfig())

	// Health endpoint without version prefix should 404
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
