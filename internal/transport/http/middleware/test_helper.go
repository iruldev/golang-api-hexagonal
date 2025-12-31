package middleware

import "github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"

// testProblemDetail represents the expected JSON response structure for testing.
// This mirrors the structure used in contract tests.
type testProblemDetail struct {
	Type             string                     `json:"type"`
	Title            string                     `json:"title"`
	Status           int                        `json:"status"`
	Detail           string                     `json:"detail"`
	Instance         string                     `json:"instance"`
	Code             string                     `json:"code"`
	RequestID        string                     `json:"request_id,omitempty"`
	TraceID          string                     `json:"trace_id,omitempty"`
	ValidationErrors []contract.ValidationError `json:"validation_errors,omitempty"`
}
