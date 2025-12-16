package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func setTracerProviderWithCleanup(t *testing.T, tp trace.TracerProvider) {
	t.Helper()
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
	})
}

func TestTracing_CreatesSpan(t *testing.T) {
	// Setup: Create a test exporter to capture spans
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer tp.Shutdown(t.Context())

	setTracerProviderWithCleanup(t, tp)

	// Setup: Create router with tracing middleware
	r := chi.NewRouter()
	r.Use(Tracing)
	r.Get("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// When: Request is made
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Then: Span is created with correct attributes
	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "/api/v1/users/{id}", span.Name)

	// Check attributes
	attrs := make(map[string]interface{})
	for _, attr := range span.Attributes {
		attrs[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, "GET", attrs["http.method"])
	assert.Equal(t, "/api/v1/users/{id}", attrs["http.route"])
	assert.Equal(t, int64(200), attrs["http.status_code"])
}

func TestTracing_PropagatesTraceContext(t *testing.T) {
	// Setup: Create a test exporter to capture spans
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer tp.Shutdown(t.Context())

	// Register test provider and propagator globally
	setTracerProviderWithCleanup(t, tp)
	prevProp := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTextMapPropagator(prevProp)
	})

	// Setup: Create router with tracing middleware
	r := chi.NewRouter()
	r.Use(Tracing)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// When: Request is made with W3C traceparent header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Then: Span should continue the existing trace
	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	// Verify the trace ID matches the incoming traceparent
	assert.Equal(t, "0af7651916cd43dd8448eb211c80319c", span.SpanContext.TraceID().String())
}

func TestTracing_StoresTraceIDInContext(t *testing.T) {
	// Setup: Create a test exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer tp.Shutdown(t.Context())

	setTracerProviderWithCleanup(t, tp)

	// Setup: Create router that captures trace ID from context
	var capturedTraceID string
	r := chi.NewRouter()
	r.Use(Tracing)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		capturedTraceID = GetTraceID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// When: Request is made
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Then: Trace ID is available in context
	assert.NotEmpty(t, capturedTraceID)
	assert.Len(t, capturedTraceID, 32) // Trace ID is 32 hex characters
}

func TestTracing_CapturesStatusCode(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		expectedStatus int64
	}{
		{
			name:           "200 OK",
			status:         http.StatusOK,
			expectedStatus: 200,
		},
		{
			name:           "201 Created",
			status:         http.StatusCreated,
			expectedStatus: 201,
		},
		{
			name:           "400 Bad Request",
			status:         http.StatusBadRequest,
			expectedStatus: 400,
		},
		{
			name:           "500 Internal Server Error",
			status:         http.StatusInternalServerError,
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			exporter := tracetest.NewInMemoryExporter()
			tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
			defer tp.Shutdown(t.Context())
			setTracerProviderWithCleanup(t, tp)

			r := chi.NewRouter()
			r.Use(Tracing)
			r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			})

			// When
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			// Then
			spans := exporter.GetSpans()
			require.Len(t, spans, 1)

			attrs := make(map[string]interface{})
			for _, attr := range spans[0].Attributes {
				attrs[string(attr.Key)] = attr.Value.AsInterface()
			}

			assert.Equal(t, tt.expectedStatus, attrs["http.status_code"])
		})
	}
}

func TestGetTraceID_ReturnsEmptyWhenNoTrace(t *testing.T) {
	// Given: A context without trace ID
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// When: GetTraceID is called
	traceID := GetTraceID(req.Context())

	// Then: Empty string is returned
	assert.Empty(t, traceID)
}

func TestTracing_FallsBackToURLPath(t *testing.T) {
	// Setup: Create a test exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer tp.Shutdown(t.Context())
	setTracerProviderWithCleanup(t, tp)

	// Setup: Use handler directly without Chi routing (simulating edge case)
	handler := Tracing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// When: Request is made without Chi route context
	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Then: Span uses URL path as name
	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "/some/path", spans[0].Name)
}

func TestGetSpanID_ReturnsEmptyWhenNoSpan(t *testing.T) {
	// Given: A context without span
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// When: GetSpanID is called
	spanID := GetSpanID(req.Context())

	// Then: Empty string is returned
	assert.Empty(t, spanID)
}

func TestTracing_StoresSpanIDInContext(t *testing.T) {
	// Setup: Create a test exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer tp.Shutdown(t.Context())

	setTracerProviderWithCleanup(t, tp)

	// Setup: Create router that captures span ID from context
	var capturedSpanID string
	r := chi.NewRouter()
	r.Use(Tracing)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		capturedSpanID = GetSpanID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// When: Request is made
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Then: Span ID is available in context with correct format
	assert.NotEmpty(t, capturedSpanID)
	assert.Len(t, capturedSpanID, 16) // Span ID is 16 hex characters (64 bits)
}

func TestGetSpanID_ReturnsCorrectFormat(t *testing.T) {
	// Setup: Create a test exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer tp.Shutdown(t.Context())

	setTracerProviderWithCleanup(t, tp)

	// Setup: Create router that captures span ID
	var capturedSpanID string
	r := chi.NewRouter()
	r.Use(Tracing)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		capturedSpanID = GetSpanID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// When: Request is made
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Then: Span ID should be exactly 16 hex characters
	assert.Len(t, capturedSpanID, 16)

	// Verify it's valid hex
	for _, c := range capturedSpanID {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"span ID should contain only hex characters, got: %c", c)
	}
}
