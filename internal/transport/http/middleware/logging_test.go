package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestRequestLogger_LogsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Create test handler that returns 200 with body
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":"test"}`))
	})

	// Create chi router with custom RequestID middleware
	r := chi.NewRouter()
	r.Use(RequestID) // Use our custom RequestID middleware
	r.Use(RequestLogger(logger))
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

	// Verify requestId is present and is 32 hex characters
	requestID, ok := logEntry["requestId"].(string)
	assert.True(t, ok, "requestId should be a string")
	assert.Len(t, requestID, 36, "requestId should be 36 characters (UUID) from custom RequestID middleware")
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

func TestRequestLogger_IncludesTraceIDWhenTracingEnabled(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Setup: Create a test tracer provider with exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer tp.Shutdown(t.Context())
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
	})

	// Create router with Tracing and RequestLogger middleware
	r := chi.NewRouter()
	r.Use(RequestID)
	r.Use(Tracing)
	r.Use(RequestLogger(logger))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// When: Request is made with tracing enabled
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Parse log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Then: traceId is present and has correct format (32 hex chars)
	// Then: traceId is present and has correct format (32 hex chars)
	traceID, ok := logEntry["traceId"].(string)
	assert.True(t, ok, "traceId should be a string")
	assert.Len(t, traceID, 32, "traceId should be 32 hex characters")

	// Then: spanId is present and has correct format (16 hex chars)
	// Then: spanId is present and has correct format (16 hex chars)
	spanID, ok := logEntry["spanId"].(string)
	assert.True(t, ok, "spanId should be a string")
	assert.Len(t, spanID, 16, "spanId should be 16 hex characters")
}

func TestRequestLogger_ExcludesTraceIDWhenTracingDisabled(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Setup: Create a noop tracer provider (disabled tracing)
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
	})

	// Create router WITHOUT Tracing middleware (simulates disabled tracing)
	r := chi.NewRouter()
	r.Use(RequestID)
	r.Use(RequestLogger(logger))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// When: Request is made without tracing
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Parse log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Then: traceId field is ABSENT (not empty string) - AC#2
	// Then: traceId field is ABSENT (not empty string) - AC#2
	_, traceIDExists := logEntry["traceId"]
	assert.False(t, traceIDExists, "traceId should be absent when tracing disabled, not empty")

	// Then: spanId field is ABSENT (not empty string) - AC#2
	// Then: spanId field is ABSENT (not empty string) - AC#2
	_, spanIDExists := logEntry["spanId"]
	assert.False(t, spanIDExists, "spanId should be absent when tracing disabled, not empty")

	// But requestId should still be present
	_, requestIDExists := logEntry["requestId"]
	assert.True(t, requestIDExists, "requestId should still be present")
}
