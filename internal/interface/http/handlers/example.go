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

// ExampleHandler demonstrates the handler pattern with context usage and error return.
//
// This handler shows how to:
// 1. Accept Context (implicitly via Request)
// 2. Return error for automatic mapping
// 3. Use wrappers for clean logic
func ExampleHandler(w http.ResponseWriter, r *http.Request) error {
	// Example: Create child span for tracing (optional)
	_, span := observability.StartSpan(r.Context(), "example-handler")
	defer span.End()

	// Access request ID from middleware (Story 3.2)
	requestID := middleware.GetRequestID(r.Context())

	// Access trace ID from OTEL middleware (Story 3.5)
	traceID := observability.GetTraceID(r.Context())

	// Demonstrate success response
	response.SuccessEnvelope(w, r.Context(), ExampleData{
		Message:   "Example handler working correctly",
		RequestID: requestID,
		TraceID:   traceID,
	})

	// Return nil to indicate success
	return nil
}
