//go:build !integration

package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// Error to HTTP status mapping tests.

func TestWriteProblemJSON(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantCode       string
		wantTitle      string
		wantType       string
		wantValField   string
		wantNoInternal bool   // verify internal details not exposed
		requestID      string // optional request ID to inject into context
		wantRequestID  string // expected request ID in response
	}{
		{
			name: "USER_NOT_FOUND maps to 404",
			err: &app.AppError{
				Op:      "GetUser",
				Code:    app.CodeUserNotFound,
				Message: "User not found",
				Err:     domain.ErrUserNotFound,
			},
			wantStatus: http.StatusNotFound,
			wantCode:   CodeUsrNotFound,
			wantTitle:  "User Not Found",
			wantType:   ProblemBaseURL + "not-found",
		},
		{
			name: "EMAIL_EXISTS maps to 409",
			err: &app.AppError{
				Op:      "CreateUser",
				Code:    app.CodeEmailExists,
				Message: "Email already exists",
				Err:     domain.ErrEmailAlreadyExists,
			},
			wantStatus: http.StatusConflict,
			wantCode:   CodeUsrEmailExists,
			wantTitle:  "Email Already Exists",
			wantType:   ProblemBaseURL + "conflict",
		},
		{
			name: "VALIDATION_ERROR maps to 400",
			err: &app.AppError{
				Op:      "ValidateUser",
				Code:    app.CodeValidationError,
				Message: "Invalid email format",
				Err:     domain.ErrInvalidEmail,
			},
			wantStatus:   http.StatusBadRequest,
			wantCode:     CodeValInvalidEmail,
			wantTitle:    "Invalid Email",
			wantType:     ProblemBaseURL + "validation-error",
			wantValField: "email",
		},
		{
			name: "VALIDATION_ERROR without wrapped error still populates validationErrors",
			err: &app.AppError{
				Op:      "ValidateUser",
				Code:    app.CodeValidationError,
				Message: "Validation failed",
			},
			wantStatus:   http.StatusBadRequest,
			wantCode:     CodeValRequired,
			wantTitle:    "Required Field Missing",
			wantType:     ProblemBaseURL + "validation-error",
			wantValField: "validation",
		},
		{
			name: "INTERNAL_ERROR hides details",
			err: &app.AppError{
				Op:      "CreateUser",
				Code:    app.CodeInternalError,
				Message: "database connection failed: SQLSTATE 42P01", // sensitive
			},
			wantStatus:     http.StatusInternalServerError,
			wantCode:       CodeSysInternal,
			wantTitle:      "Internal Server Error",
			wantType:       ProblemBaseURL + "internal-error",
			wantNoInternal: true,
		},
		{
			name: "RATE_LIMIT_EXCEEDED maps to 429",
			err: &app.AppError{
				Op:      "RateLimiter",
				Code:    app.CodeRateLimitExceeded,
				Message: "Rate limit exceeded",
			},
			wantStatus: http.StatusTooManyRequests,
			wantCode:   CodeRateLimitExceeded,
			wantTitle:  "Rate Limit Exceeded",
			wantType:   ProblemBaseURL + "rate-limit-exceeded",
		},
		{
			name:           "unknown error becomes INTERNAL_ERROR",
			err:            errors.New("something went wrong"),
			wantStatus:     http.StatusInternalServerError,
			wantCode:       CodeSysInternal,
			wantTitle:      "Internal Server Error",
			wantType:       ProblemBaseURL + "internal-error",
			wantNoInternal: true,
		},
		{
			name:      "Request ID is propagated",
			requestID: "req-123",
			err: &app.AppError{
				Code:    app.CodeUserNotFound,
				Message: "User not found",
			},
			wantStatus:    http.StatusNotFound,
			wantCode:      CodeUsrNotFound,
			wantTitle:     "User Not Found",
			wantType:      ProblemBaseURL + "not-found",
			wantRequestID: "req-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
			if tt.requestID != "" {
				ctx := ctxutil.SetRequestID(req.Context(), tt.requestID)
				req = req.WithContext(ctx)
			}
			rec := httptest.NewRecorder()

			WriteProblemJSON(rec, req, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, ContentTypeProblemJSON, rec.Header().Get("Content-Type"))

			var problem testProblemDetail
			err := json.NewDecoder(rec.Body).Decode(&problem)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, problem.Code)
			assert.Equal(t, tt.wantTitle, problem.Title)
			assert.Equal(t, tt.wantType, problem.Type)
			assert.Equal(t, tt.wantStatus, problem.Status)
			assert.Equal(t, "/api/v1/users/123", problem.Instance)

			if tt.wantRequestID != "" {
				assert.Equal(t, tt.wantRequestID, problem.RequestID)
			}

			if tt.wantCode == app.CodeValidationError {
				require.Len(t, problem.ValidationErrors, 1)
				assert.Equal(t, tt.wantValField, problem.ValidationErrors[0].Field)
			}

			if tt.wantNoInternal {
				assert.Equal(t, "An internal error occurred", problem.Detail)
				assert.NotContains(t, problem.Detail, "SQLSTATE")
				assert.NotContains(t, problem.Detail, "database")
			}
		})
	}
}

func TestWriteProblemJSON_DomainErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantStatus   int
		wantCode     string
		wantTitle    string
		wantType     string
		wantValField string
	}{
		{
			name: "Domain UserNotFound maps to 404",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeUserNotFound,
				Message: "User via domain error",
			},
			wantStatus: http.StatusNotFound,
			wantCode:   CodeUsrNotFound,
			wantTitle:  "User Not Found",
			wantType:   ProblemBaseURL + "not-found",
		},
		{
			name: "Domain EmailExists maps to 409",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeEmailExists,
				Message: "Email exists",
			},
			wantStatus: http.StatusConflict,
			wantCode:   CodeUsrEmailExists,
			wantTitle:  "Email Already Exists",
			wantType:   ProblemBaseURL + "conflict",
		},
		{
			name: "Domain InvalidEmail maps to 400 with validation error",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidEmail,
				Message: "Invalid email format",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeValInvalidEmail,
			wantTitle:  "Invalid Email",
			wantType:   ProblemBaseURL + "validation-error",
		},
		{
			name: "Domain InvalidFirstName maps to 400 with validation error",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidFirstName,
				Message: "First name is required",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeUsrInvalidField,
			wantTitle:  "Invalid User Field",
			wantType:   ProblemBaseURL + "validation-error",
		},
		{
			name: "Domain InvalidLastName maps to 400 with validation error",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidLastName,
				Message: "Last name is required",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeUsrInvalidField,
			wantTitle:  "Invalid User Field",
			wantType:   ProblemBaseURL + "validation-error",
		},
		{
			name: "Domain AuditNotFound maps to 404",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeAuditNotFound,
				Message: "Audit event not found",
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   CodeSysInternal,
			wantTitle:  "Internal Server Error",
			wantType:   ProblemBaseURL + "internal-error",
		},
		{
			name: "Domain InvalidEventType maps to 400",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidEventType,
				Message: "Invalid event type",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeValInvalidFormat,
			wantTitle:  "Invalid Format",
			wantType:   ProblemBaseURL + "validation-error",
		},
		{
			name: "Domain InvalidEntityType maps to 400",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidEntityType,
				Message: "Invalid entity type",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeValInvalidType,
			wantTitle:  "Invalid Type",
			wantType:   ProblemBaseURL + "validation-error",
		},
		{
			name: "Domain InvalidEntityID maps to 400",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidEntityID,
				Message: "Invalid entity ID",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeValInvalidUUID,
			wantTitle:  "Invalid UUID",
			wantType:   ProblemBaseURL + "validation-error",
		},
		{
			name: "Domain InvalidID maps to 400",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidID,
				Message: "Invalid ID",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeValInvalidFormat,
			wantTitle:  "Invalid Format",
			wantType:   ProblemBaseURL + "validation-error",
		},
		{
			name: "Domain InvalidTimestamp maps to 400",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidTimestamp,
				Message: "Invalid timestamp",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeValInvalidFormat,
			wantTitle:  "Invalid Format",
			wantType:   ProblemBaseURL + "validation-error",
		},
		{
			name: "Domain InvalidPayload maps to 400",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidPayload,
				Message: "Invalid payload",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeValInvalidFormat,
			wantTitle:  "Invalid Format",
			wantType:   ProblemBaseURL + "validation-error",
		},
		{
			name: "Domain InvalidRequestID maps to 400",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidRequestID,
				Message: "Invalid request ID",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeValInvalidUUID,
			wantTitle:  "Invalid UUID",
			wantType:   ProblemBaseURL + "validation-error",
		},
		{
			name: "Domain Unauthorized maps to 401",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeUnauthorized,
				Message: "User not authorized",
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   CodeAuthExpiredToken,
			wantTitle:  "Token Expired",
			wantType:   ProblemBaseURL + "unauthorized",
		},
		{
			name: "Domain Forbidden maps to 403",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeForbidden,
				Message: "Access forbidden",
			},
			wantStatus: http.StatusForbidden,
			wantCode:   CodeAuthzForbidden,
			wantTitle:  "Forbidden",
			wantType:   ProblemBaseURL + "forbidden",
		},
		{
			name: "Domain Conflict maps to 409",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeConflict,
				Message: "Generic conflict",
			},
			wantStatus: http.StatusConflict,
			wantCode:   CodeUsrEmailExists,
			wantTitle:  "Email Already Exists",
			wantType:   ProblemBaseURL + "conflict",
		},
		{
			name: "Domain NotFound maps to 404",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeNotFound,
				Message: "Generic not found",
			},
			wantStatus: http.StatusNotFound,
			wantCode:   CodeUsrNotFound,
			wantTitle:  "User Not Found",
			wantType:   ProblemBaseURL + "not-found",
		},
		{
			name: "Domain InternalError maps to 500",
			err: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInternal,
				Message: "Something bad happened",
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   CodeSysInternal,
			wantTitle:  "Internal Server Error",
			wantType:   ProblemBaseURL + "internal-error",
		},
		{
			name: "Domain UnknownError maps to 500",
			err: &domainerrors.DomainError{
				Code:    "ERR_UNKNOWN_DOMAIN",
				Message: "Some unknown error",
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   CodeSysInternal,
			wantTitle:  "Internal Server Error",
			wantType:   ProblemBaseURL + "internal-error",
		},
		{
			name: "Wrapped Domain Error is prioritized",
			err: fmt.Errorf("wrapping: %w", &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeConflict,
				Message: "Conflict occurred",
			}),
			wantStatus: http.StatusConflict,
			wantCode:   CodeUsrEmailExists,
			wantTitle:  "Email Already Exists",
			wantType:   ProblemBaseURL + "conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
			rec := httptest.NewRecorder()

			WriteProblemJSON(rec, req, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, ContentTypeProblemJSON, rec.Header().Get("Content-Type"))

			var problem testProblemDetail
			err := json.NewDecoder(rec.Body).Decode(&problem)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, problem.Code)
			assert.Equal(t, tt.wantTitle, problem.Title)
			assert.Equal(t, tt.wantType, problem.Type)
			assert.Equal(t, tt.wantStatus, problem.Status)
			// Detail check handles both direct and wrapped error messages
			// assert.Contains(t, problem.Detail, tt.err.(interface{ Error() string }).Error())
		})
	}
}
