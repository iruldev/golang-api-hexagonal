package contract_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/resilience"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

// mockCodeError implements codeGetter interface for testing
type mockCodeError struct {
	code string
}

func (e *mockCodeError) Error() string {
	return "mock error"
}

func (e *mockCodeError) GetCode() string {
	return e.code
}

func TestMapErrorToCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		// Nil error
		{
			name:     "nil error returns SYS-001",
			err:      nil,
			wantCode: contract.CodeSysInternal,
		},

		// Domain errors
		{
			name: "domain user not found",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeUserNotFound,
				Message: "user not found",
			},
			wantCode: contract.CodeUsrNotFound,
		},
		{
			name: "domain email exists",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeEmailExists,
				Message: "email exists",
			},
			wantCode: contract.CodeUsrEmailExists,
		},
		{
			name: "domain invalid email",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidEmail,
				Message: "invalid email",
			},
			wantCode: contract.CodeValInvalidEmail,
		},

		// App errors
		{
			name: "app unauthorized",
			err: &app.AppError{
				Code:    app.CodeUnauthorized,
				Message: "unauthorized",
			},
			wantCode: contract.CodeAuthExpiredToken, // ERR_UNAUTHORIZED maps to AUTH-001
		},
		{
			name: "app forbidden",
			err: &app.AppError{
				Code:    app.CodeForbidden,
				Message: "forbidden",
			},
			wantCode: contract.CodeAuthzForbidden,
		},
		{
			name: "app validation error",
			err: &app.AppError{
				Code:    app.CodeValidationError,
				Message: "validation failed",
			},
			wantCode: contract.CodeValRequired,
		},
		{
			name: "app internal error",
			err: &app.AppError{
				Code:    app.CodeInternalError,
				Message: "internal error",
			},
			wantCode: contract.CodeSysInternal,
		},

		// Resilience errors
		{
			name: "resilience circuit open",
			err: &resilience.ResilienceError{
				Code:    resilience.ErrCodeCircuitOpen,
				Message: "circuit open",
			},
			wantCode: contract.CodeResCircuitOpen,
		},
		{
			name: "resilience bulkhead full",
			err: &resilience.ResilienceError{
				Code:    resilience.ErrCodeBulkheadFull,
				Message: "bulkhead full",
			},
			wantCode: contract.CodeResBulkheadFull,
		},
		{
			name: "resilience timeout exceeded",
			err: &resilience.ResilienceError{
				Code:    resilience.ErrCodeTimeoutExceeded,
				Message: "timeout",
			},
			wantCode: contract.CodeResTimeoutExceeded,
		},
		{
			name: "resilience max retries exceeded",
			err: &resilience.ResilienceError{
				Code:    resilience.ErrCodeMaxRetriesExceeded,
				Message: "max retries",
			},
			wantCode: contract.CodeResMaxRetriesExceeded,
		},
		{
			name: "resilience unknown code",
			err: &resilience.ResilienceError{
				Code:    "RES-999",
				Message: "unknown resilience error",
			},
			wantCode: contract.CodeSysUnavailable,
		},

		// Generic codeGetter
		{
			name: "generic code getter non-RES",
			err: &mockCodeError{
				code: "CUSTOM-001",
			},
			wantCode: contract.CodeSysInternal,
		},

		// Unknown error
		{
			name:     "unknown error returns SYS-001",
			err:      fmt.Errorf("some random error"),
			wantCode: contract.CodeSysInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contract.MapErrorToCode(tt.err)
			if got != tt.wantCode {
				t.Errorf("MapErrorToCode() = %q, want %q", got, tt.wantCode)
			}
		})
	}
}

func TestMapErrorToCode_WrappedErrors(t *testing.T) {
	// Test that wrapped errors are properly unwrapped
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{
			name: "wrapped domain error",
			err: fmt.Errorf("operation failed: %w", &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeUserNotFound,
				Message: "user not found",
			}),
			wantCode: contract.CodeUsrNotFound,
		},
		{
			name: "wrapped app error",
			err: fmt.Errorf("service error: %w", &app.AppError{
				Code:    app.CodeForbidden,
				Message: "access denied",
			}),
			wantCode: contract.CodeAuthzForbidden,
		},
		{
			name: "wrapped resilience error",
			err: fmt.Errorf("call failed: %w", &resilience.ResilienceError{
				Code:    resilience.ErrCodeCircuitOpen,
				Message: "circuit open",
			}),
			wantCode: contract.CodeResCircuitOpen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contract.MapErrorToCode(tt.err)
			if got != tt.wantCode {
				t.Errorf("MapErrorToCode() = %q, want %q", got, tt.wantCode)
			}
		})
	}
}

func TestMapErrorToHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name: "domain not found → 404",
			err: &domainerrors.DomainError{
				Code: domainerrors.ErrCodeUserNotFound,
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "app forbidden → 403",
			err: &app.AppError{
				Code: app.CodeForbidden,
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "resilience circuit open → 503",
			err: &resilience.ResilienceError{
				Code: resilience.ErrCodeCircuitOpen,
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "unknown error → 500",
			err:        fmt.Errorf("unknown"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contract.MapErrorToHTTPStatus(tt.err)
			if got != tt.wantStatus {
				t.Errorf("MapErrorToHTTPStatus() = %d, want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestIsClientError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "validation error is client error",
			err: &app.AppError{
				Code: app.CodeValidationError,
			},
			want: true,
		},
		{
			name: "not found is client error",
			err: &domainerrors.DomainError{
				Code: domainerrors.ErrCodeUserNotFound,
			},
			want: true,
		},
		{
			name: "internal error is not client error",
			err: &app.AppError{
				Code: app.CodeInternalError,
			},
			want: false,
		},
		{
			name: "resilience error is not client error",
			err: &resilience.ResilienceError{
				Code: resilience.ErrCodeCircuitOpen,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contract.IsClientError(tt.err)
			if got != tt.want {
				t.Errorf("IsClientError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsServerError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "internal error is server error",
			err: &app.AppError{
				Code: app.CodeInternalError,
			},
			want: true,
		},
		{
			name: "resilience error is server error",
			err: &resilience.ResilienceError{
				Code: resilience.ErrCodeCircuitOpen,
			},
			want: true,
		},
		{
			name: "validation error is not server error",
			err: &app.AppError{
				Code: app.CodeValidationError,
			},
			want: false,
		},
		{
			name: "not found is not server error",
			err: &domainerrors.DomainError{
				Code: domainerrors.ErrCodeUserNotFound,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contract.IsServerError(tt.err)
			if got != tt.want {
				t.Errorf("IsServerError() = %v, want %v", got, tt.want)
			}
		})
	}
}
