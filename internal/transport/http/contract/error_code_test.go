//go:build !integration

package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
)

// Error code registry and safe detail generation tests.

// testProblemDetail represents the expected JSON response structure for testing.
type testProblemDetail struct {
	Type             string            `json:"type"`
	Title            string            `json:"title"`
	Status           int               `json:"status"`
	Detail           string            `json:"detail"`
	Instance         string            `json:"instance"`
	Code             string            `json:"code"`
	RequestID        string            `json:"request_id,omitempty"`
	TraceID          string            `json:"trace_id,omitempty"`
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
}

func TestSafeDetail(t *testing.T) {
	tests := []struct {
		name       string
		appErr     *app.AppError
		wantDetail string
	}{
		{
			name: "4xx shows original message",
			appErr: &app.AppError{
				Code:    app.CodeUserNotFound,
				Message: "User with ID 123 not found",
			},
			wantDetail: "User with ID 123 not found",
		},
		{
			name: "5xx hides internal details",
			appErr: &app.AppError{
				Code:    app.CodeInternalError,
				Message: "database query failed: connection refused",
			},
			wantDetail: "An internal error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeAppErrorDetail(tt.appErr)
			assert.Equal(t, tt.wantDetail, got)
		})
	}
}
