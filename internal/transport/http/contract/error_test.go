//go:build !integration

package contract

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

func TestWriteProblemJSON(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantStatus     int
		wantCode       string
		wantTitle      string
		wantType       string
		wantValField   string
		wantNoInternal bool // verify internal details not exposed
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
			wantCode:   app.CodeUserNotFound,
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
			wantCode:   app.CodeEmailExists,
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
			wantCode:     app.CodeValidationError,
			wantTitle:    "Validation Error",
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
			wantCode:     app.CodeValidationError,
			wantTitle:    "Validation Error",
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
			wantCode:       app.CodeInternalError,
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
			wantCode:   app.CodeRateLimitExceeded,
			wantTitle:  "Too Many Requests",
			wantType:   ProblemBaseURL + "rate-limit-exceeded",
		},
		{
			name:           "unknown error becomes INTERNAL_ERROR",
			err:            errors.New("something went wrong"),
			wantStatus:     http.StatusInternalServerError,
			wantCode:       app.CodeInternalError,
			wantTitle:      "Internal Server Error",
			wantType:       ProblemBaseURL + "internal-error",
			wantNoInternal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
			rec := httptest.NewRecorder()

			WriteProblemJSON(rec, req, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

			var problem ProblemDetail
			err := json.NewDecoder(rec.Body).Decode(&problem)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, problem.Code)
			assert.Equal(t, tt.wantTitle, problem.Title)
			assert.Equal(t, tt.wantType, problem.Type)
			assert.Equal(t, tt.wantStatus, problem.Status)
			assert.Equal(t, "/api/v1/users/123", problem.Instance)

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

func TestWriteValidationError(t *testing.T) {
	validationErrors := []ValidationError{
		{Field: "email", Message: "must be a valid email address"},
		{Field: "firstName", Message: "must not be empty"},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	WriteValidationError(rec, req, validationErrors)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var problem ProblemDetail
	err := json.NewDecoder(rec.Body).Decode(&problem)
	require.NoError(t, err)

	assert.Equal(t, app.CodeValidationError, problem.Code)
	assert.Equal(t, "Validation Error", problem.Title)
	assert.Equal(t, ProblemBaseURL+"validation-error", problem.Type)
	assert.Equal(t, http.StatusBadRequest, problem.Status)
	assert.Equal(t, "/api/v1/users", problem.Instance)
	assert.Equal(t, "One or more fields failed validation", problem.Detail)

	require.Len(t, problem.ValidationErrors, 2)
	assert.Equal(t, "email", problem.ValidationErrors[0].Field)
	assert.Equal(t, "must be a valid email address", problem.ValidationErrors[0].Message)
	assert.Equal(t, "firstName", problem.ValidationErrors[1].Field)
	assert.Equal(t, "must not be empty", problem.ValidationErrors[1].Message)
}

func TestNewValidationProblem(t *testing.T) {
	validationErrors := []ValidationError{
		{Field: "email", Message: "required"},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	problem := NewValidationProblem(req, validationErrors)

	assert.NotNil(t, problem)
	assert.Equal(t, ProblemBaseURL+"validation-error", problem.Type)
	assert.Equal(t, "Validation Error", problem.Title)
	assert.Equal(t, http.StatusBadRequest, problem.Status)
	assert.Equal(t, app.CodeValidationError, problem.Code)
	assert.Equal(t, "/api/v1/users", problem.Instance)
	assert.Len(t, problem.ValidationErrors, 1)
}

func TestMapCodeToStatus(t *testing.T) {
	tests := []struct {
		code       string
		wantStatus int
	}{
		{app.CodeUserNotFound, http.StatusNotFound},
		{app.CodeEmailExists, http.StatusConflict},
		{app.CodeValidationError, http.StatusBadRequest},
		{app.CodeUnauthorized, http.StatusUnauthorized},
		{app.CodeForbidden, http.StatusForbidden},
		{app.CodeRateLimitExceeded, http.StatusTooManyRequests},
		{app.CodeInternalError, http.StatusInternalServerError},
		{"UNKNOWN_CODE", http.StatusInternalServerError},
		{"", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := mapCodeToStatus(tt.code)
			assert.Equal(t, tt.wantStatus, got)
		})
	}
}

func TestProblemTypeURL(t *testing.T) {
	tests := []struct {
		slug    string
		wantURL string
	}{
		{ProblemTypeNotFoundSlug, ProblemBaseURL + "not-found"},
		{ProblemTypeConflictSlug, ProblemBaseURL + "conflict"},
		{ProblemTypeValidationErrorSlug, ProblemBaseURL + "validation-error"},
		{ProblemTypeRateLimitSlug, ProblemBaseURL + "rate-limit-exceeded"},
		{ProblemTypeInternalErrorSlug, ProblemBaseURL + "internal-error"},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			got := problemTypeURL(tt.slug)
			assert.Equal(t, tt.wantURL, got)
		})
	}
}

func TestCodeToTitle(t *testing.T) {
	tests := []struct {
		code      string
		wantTitle string
	}{
		{app.CodeUserNotFound, "User Not Found"},
		{app.CodeEmailExists, "Email Already Exists"},
		{app.CodeValidationError, "Validation Error"},
		{app.CodeValidationError, "Validation Error"},
		{app.CodeRateLimitExceeded, "Too Many Requests"},
		{app.CodeInternalError, "Internal Server Error"},
		{"UNKNOWN_CODE", "Internal Server Error"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := codeToTitle(tt.code)
			assert.Equal(t, tt.wantTitle, got)
		})
	}
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
			got := safeDetail(tt.appErr)
			assert.Equal(t, tt.wantDetail, got)
		})
	}
}

func TestProblemDetailJSON(t *testing.T) {
	// Test that ProblemDetail serializes to correct JSON format
	problem := ProblemDetail{
		Type:     ProblemBaseURL + "validation-error",
		Title:    "Validation Error",
		Status:   400,
		Detail:   "One or more fields failed validation",
		Instance: "/api/v1/users",
		Code:     app.CodeValidationError,
		ValidationErrors: []ValidationError{
			{Field: "email", Message: "required"},
		},
	}

	data, err := json.Marshal(problem)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Verify camelCase keys
	assert.Contains(t, parsed, "type")
	assert.Contains(t, parsed, "title")
	assert.Contains(t, parsed, "status")
	assert.Contains(t, parsed, "detail")
	assert.Contains(t, parsed, "instance")
	assert.Contains(t, parsed, "code")
	assert.Contains(t, parsed, "validationErrors")

	// Verify values
	assert.Equal(t, ProblemBaseURL+"validation-error", parsed["type"])
	assert.Equal(t, float64(400), parsed["status"])
}

func TestProblemDetailOmitEmpty(t *testing.T) {
	// Test that ValidationErrors is omitted when empty
	problem := ProblemDetail{
		Type:     ProblemBaseURL + "not-found",
		Title:    "Not Found",
		Status:   404,
		Detail:   "Resource not found",
		Instance: "/api/v1/users/123",
		Code:     app.CodeUserNotFound,
		// ValidationErrors is empty/nil
	}

	data, err := json.Marshal(problem)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// validationErrors should be omitted
	assert.NotContains(t, parsed, "validationErrors")
}
