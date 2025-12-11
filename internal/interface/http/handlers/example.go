package handlers

import (
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// ExampleData represents the example endpoint data.
type ExampleData struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	TraceID   string `json:"trace_id,omitempty"`
}

// ExampleHandler demonstrates the handler pattern with context usage.
//
// This handler shows how to:
// - Access request ID from RequestID middleware
// - Access trace ID from OTEL middleware
// - Create child spans for tracing operations
// - Use response envelope pattern (Story 3.7)
//
// Endpoint: GET /api/v1/example
// Response: {"success": true, "data": {...}}
func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	// Access request ID from middleware (Story 3.2)
	requestID := middleware.GetRequestID(r.Context())

	// Access trace ID from OTEL middleware (Story 3.5)
	traceID := observability.GetTraceID(r.Context())

	// Example: Create child span for tracing (optional)
	// Use the span context for downstream operations like database or API calls
	_, span := observability.StartSpan(r.Context(), "example-handler")
	defer span.End()

	// Build response data and send with envelope pattern (Story 3.7)
	response.Success(w, ExampleData{
		Message:   "Example handler working correctly",
		RequestID: requestID,
		TraceID:   traceID,
	})
}
