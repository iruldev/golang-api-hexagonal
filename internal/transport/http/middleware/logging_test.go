package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogger_LogsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Create test handler that returns 200 with body
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":"test"}`))
	})

	// Create chi router with our middleware
	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(RequestLogger(logger))
	r.Use(chiMiddleware.RequestID)
	r.Get("/test", testHandler)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Parse log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err, "log output should be valid JSON")

	// Verify fields
	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/test", logEntry["route"])
	assert.Equal(t, float64(http.StatusOK), logEntry["status"])
	assert.NotNil(t, logEntry["duration_ms"])
	assert.Equal(t, float64(len(`{"data":"test"}`)), logEntry["bytes"])
	assert.NotEmpty(t, logEntry["requestId"], "requestId should be present from chi RequestID middleware")
}

func TestRequestLogger_CapturesErrorStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Handler returns 500
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	})

	r := chi.NewRouter()
	r.Use(RequestLogger(logger))
	r.Get("/error", testHandler)

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, float64(http.StatusInternalServerError), logEntry["status"])
	assert.Equal(t, float64(len("error")), logEntry["bytes"])
}

func TestRequestLogger_CapturesRoutePattern(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := chi.NewRouter()
	r.Use(RequestLogger(logger))
	r.Get("/users/{id}", testHandler)

	// Make request with specific ID
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Should log the pattern, not the actual path
	assert.Equal(t, "/users/{id}", logEntry["route"])
}

func TestRequestLogger_DurationIsPositive(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := chi.NewRouter()
	r.Use(RequestLogger(logger))
	r.Get("/fast", testHandler)

	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Duration should be >= 0
	duration, ok := logEntry["duration_ms"].(float64)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, duration, float64(0))
}

func TestRequestLogger_NoRouteUsesPath(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Direct middleware call without chi to simulate empty route pattern
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Use chi router but request a path that doesn't match any route
	r := chi.NewRouter()
	r.Use(RequestLogger(logger))
	r.Get("/known", testHandler)

	// Request unknown path (404)
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// For 404, route pattern is empty so we fall back to path
	assert.Equal(t, "/unknown", logEntry["route"])
}
