//go:build !integration

package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// RFC 7807 response format tests.

func TestWriteValidationError(t *testing.T) {
	validationErrors := []ValidationError{
		{Field: "email", Message: "must be a valid email address"},
		{Field: "firstName", Message: "must not be empty"},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	WriteValidationError(rec, req, validationErrors)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, ContentTypeProblemJSON, rec.Header().Get("Content-Type"))

	var problem testProblemDetail
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

func TestProblemDetailJSON(t *testing.T) {
	// Test that ProblemDetail serializes to correct JSON format
	problem := testProblemDetail{
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
	assert.Contains(t, parsed, "validation_errors")

	// Verify values
	assert.Equal(t, ProblemBaseURL+"validation-error", parsed["type"])
	assert.Equal(t, float64(400), parsed["status"])
}

func TestProblemDetailOmitEmpty(t *testing.T) {
	// Test that ValidationErrors is omitted when empty
	problem := testProblemDetail{
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
	assert.NotContains(t, parsed, "validation_errors")
}

func TestWriteProblemJSON_IncludesRequestID(t *testing.T) {
	// Create request with request ID in context
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	ctx := ctxutil.SetRequestID(req.Context(), "test-request-id-abc")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	err := &app.AppError{
		Op:      "GetUser",
		Code:    app.CodeUserNotFound,
		Message: "User not found",
		Err:     domain.ErrUserNotFound,
	}

	WriteProblemJSON(rec, req, err)

	var problem testProblemDetail
	decodeErr := json.NewDecoder(rec.Body).Decode(&problem)
	require.NoError(t, decodeErr)

	assert.Equal(t, "test-request-id-abc", problem.RequestID, "request_id should be present in error response")
}

func TestWriteProblemJSON_NoRequestID_GracefulDegradation(t *testing.T) {
	// Create request without request ID in context
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	rec := httptest.NewRecorder()

	err := &app.AppError{
		Op:      "GetUser",
		Code:    app.CodeUserNotFound,
		Message: "User not found",
	}

	WriteProblemJSON(rec, req, err)

	var problem testProblemDetail
	decodeErr := json.NewDecoder(rec.Body).Decode(&problem)
	require.NoError(t, decodeErr)

	// Empty request_id should be omitted from JSON (omitempty)
	assert.Empty(t, problem.RequestID, "request_id should be empty when not in context")
}

func TestNewValidationProblem_IncludesRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	ctx := ctxutil.SetRequestID(req.Context(), "validation-request-123")
	req = req.WithContext(ctx)

	validationErrors := []ValidationError{
		{Field: "email", Message: "required"},
	}

	problem := NewValidationProblem(req, validationErrors)

	assert.Equal(t, "validation-request-123", problem.RequestID, "request_id should be present in validation error")
}

func TestWriteProblemJSON_IncludesTraceID(t *testing.T) {
	// Create request with trace ID in context
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	ctx := ctxutil.SetTraceID(req.Context(), "abc123def456789012345678901234ab")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	err := &app.AppError{
		Op:      "GetUser",
		Code:    app.CodeUserNotFound,
		Message: "User not found",
		Err:     domain.ErrUserNotFound,
	}

	WriteProblemJSON(rec, req, err)

	var problem testProblemDetail
	decodeErr := json.NewDecoder(rec.Body).Decode(&problem)
	require.NoError(t, decodeErr)

	assert.Equal(t, "abc123def456789012345678901234ab", problem.TraceID, "trace_id should be present in error response")
}

func TestWriteProblemJSON_ZeroTraceID_Omitted(t *testing.T) {
	// Create request with zero trace ID (tracing disabled/not started)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	ctx := ctxutil.SetTraceID(req.Context(), ctxutil.EmptyTraceID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	err := &app.AppError{
		Op:      "GetUser",
		Code:    app.CodeUserNotFound,
		Message: "User not found",
	}

	WriteProblemJSON(rec, req, err)

	var problem testProblemDetail
	decodeErr := json.NewDecoder(rec.Body).Decode(&problem)
	require.NoError(t, decodeErr)

	// Zero trace_id should be omitted (graceful degradation)
	assert.Empty(t, problem.TraceID, "zero trace_id should be omitted")
}

func TestWriteProblemJSON_BothRequestIDAndTraceID(t *testing.T) {
	// Create request with both request_id and trace_id
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	ctx := ctxutil.SetRequestID(req.Context(), "req-abc-123")
	ctx = ctxutil.SetTraceID(ctx, "trace-xyz-789012345678901234")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	err := &app.AppError{
		Op:      "GetUser",
		Code:    app.CodeUserNotFound,
		Message: "User not found",
	}

	WriteProblemJSON(rec, req, err)

	var problem testProblemDetail
	decodeErr := json.NewDecoder(rec.Body).Decode(&problem)
	require.NoError(t, decodeErr)

	assert.Equal(t, "req-abc-123", problem.RequestID, "request_id should be present")
	assert.Equal(t, "trace-xyz-789012345678901234", problem.TraceID, "trace_id should be present")
}

func TestNewValidationProblem_IncludesTraceID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
	ctx := ctxutil.SetTraceID(req.Context(), "validation-trace-456")
	req = req.WithContext(ctx)

	validationErrors := []ValidationError{
		{Field: "email", Message: "required"},
	}

	problem := NewValidationProblem(req, validationErrors)

	assert.Equal(t, "validation-trace-456", problem.TraceID, "trace_id should be present in validation error")
}
