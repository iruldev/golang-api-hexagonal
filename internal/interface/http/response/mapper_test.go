package response

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

func TestMapError_NotFound(t *testing.T) {
	status, code := MapError(domain.ErrNotFound)

	if status != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", status)
	}
	if code != CodeNotFound {
		t.Errorf("Expected code %s, got %s", CodeNotFound, code)
	}
}

func TestMapError_Validation(t *testing.T) {
	status, code := MapError(domain.ErrValidation)

	if status != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", status)
	}
	// Use domainerrors.CodeValidationError (VALIDATION_ERROR) instead of legacy VALIDATION_FAILED
	if code != domainerrors.CodeValidationError {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeValidationError, code)
	}
}

func TestMapError_Unauthorized(t *testing.T) {
	status, code := MapError(domain.ErrUnauthorized)

	if status != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", status)
	}
	if code != CodeUnauthorized {
		t.Errorf("Expected code %s, got %s", CodeUnauthorized, code)
	}
}

func TestMapError_Forbidden(t *testing.T) {
	status, code := MapError(domain.ErrForbidden)

	if status != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", status)
	}
	if code != CodeForbidden {
		t.Errorf("Expected code %s, got %s", CodeForbidden, code)
	}
}

func TestMapError_Conflict(t *testing.T) {
	status, code := MapError(domain.ErrConflict)

	if status != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", status)
	}
	if code != CodeConflict {
		t.Errorf("Expected code %s, got %s", CodeConflict, code)
	}
}

func TestMapError_Internal(t *testing.T) {
	status, code := MapError(domain.ErrInternal)

	if status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", status)
	}
	if code != domainerrors.CodeInternalError {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeInternalError, code)
	}
}

func TestMapError_Unknown(t *testing.T) {
	// Use an error that doesn't match any domain error
	unknownErr := errors.New("database connection timeout")
	status, code := MapError(unknownErr)

	if status != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for unknown error, got %d", status)
	}
	if code != domainerrors.CodeInternalError {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeInternalError, code)
	}
}

func TestMapError_WrappedNotFound(t *testing.T) {
	wrapped := domain.WrapError(domain.ErrNotFound, "user not found")
	status, code := MapError(wrapped)

	if status != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", status)
	}
	if code != CodeNotFound {
		t.Errorf("Expected code %s, got %s", CodeNotFound, code)
	}
}

// Tests for DomainError mapping (Story 2.2)

func TestMapError_DomainError_NotFound(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeNotFound, "note not found")
	status, code := MapError(err)

	if status != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", status)
	}
	if code != domainerrors.CodeNotFound {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeNotFound, code)
	}
}

func TestMapError_DomainError_ValidationError(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeValidationError, "invalid input")
	status, code := MapError(err)

	if status != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", status)
	}
	if code != domainerrors.CodeValidationError {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeValidationError, code)
	}
}

func TestMapError_DomainError_Unauthorized(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeUnauthorized, "not authorized")
	status, code := MapError(err)

	if status != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", status)
	}
	if code != domainerrors.CodeUnauthorized {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeUnauthorized, code)
	}
}

func TestMapError_DomainError_Forbidden(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeForbidden, "access denied")
	status, code := MapError(err)

	if status != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", status)
	}
	if code != domainerrors.CodeForbidden {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeForbidden, code)
	}
}

func TestMapError_DomainError_Conflict(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeConflict, "resource conflict")
	status, code := MapError(err)

	if status != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", status)
	}
	if code != domainerrors.CodeConflict {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeConflict, code)
	}
}

func TestMapError_DomainError_InternalError(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeInternalError, "internal error")
	status, code := MapError(err)

	if status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", status)
	}
	if code != domainerrors.CodeInternalError {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeInternalError, code)
	}
}

func TestMapError_DomainError_Timeout(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeTimeout, "operation timed out")
	status, code := MapError(err)

	if status != http.StatusGatewayTimeout {
		t.Errorf("Expected status 504, got %d", status)
	}
	if code != domainerrors.CodeTimeout {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeTimeout, code)
	}
}

func TestMapError_DomainError_RateLimitExceeded(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeRateLimitExceeded, "rate limit exceeded")
	status, code := MapError(err)

	if status != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", status)
	}
	if code != domainerrors.CodeRateLimitExceeded {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeRateLimitExceeded, code)
	}
}

func TestMapError_DomainError_BadRequest(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeBadRequest, "bad request")
	status, code := MapError(err)

	if status != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", status)
	}
	if code != domainerrors.CodeBadRequest {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeBadRequest, code)
	}
}

func TestMapError_DomainError_UnknownCode(t *testing.T) {
	// Manually construct struct to bypass NewDomain validation
	// This simulates a scenario where a code exists in registry but not in mapper,
	// or strict validation is bypassed.
	err := &domainerrors.DomainError{
		Code:    "UNKNOWN_CODE",
		Message: "unknown error",
	}
	status, code := MapError(err)

	if status != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for unknown code, got %d", status)
	}
	if code != "UNKNOWN_CODE" {
		t.Errorf("Expected code UNKNOWN_CODE, got %s", code)
	}
}

func TestMapErrorWithHint_DomainError(t *testing.T) {
	err := domainerrors.NewDomainWithHint(
		domainerrors.CodeValidationError,
		"invalid email format",
		"email must be like user@example.com",
	)

	status, code, message, hint := MapErrorWithHint(err)

	if status != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", status)
	}
	if code != domainerrors.CodeValidationError {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeValidationError, code)
	}
	if message != "invalid email format" {
		t.Errorf("Expected message 'invalid email format', got %s", message)
	}
	if hint != "email must be like user@example.com" {
		t.Errorf("Expected hint 'email must be like user@example.com', got %s", hint)
	}
}

func TestMapErrorWithHint_NonDomainError(t *testing.T) {
	err := domain.WrapError(domain.ErrNotFound, "user not found")

	status, code, message, hint := MapErrorWithHint(err)

	if status != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", status)
	}
	if code != CodeNotFound {
		t.Errorf("Expected code %s, got %s", CodeNotFound, code)
	}
	if message != "user not found" {
		t.Errorf("Expected message 'user not found', got %s", message)
	}
	if hint != "" {
		t.Errorf("Expected empty hint, got %s", hint)
	}
}

func TestHandleError_NotFound(t *testing.T) {
	rr := httptest.NewRecorder()

	HandleError(rr, domain.ErrNotFound)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if response.Success {
		t.Error("Expected success to be false")
	}
	if response.Error.Code != CodeNotFound {
		t.Errorf("Expected code %s, got %s", CodeNotFound, response.Error.Code)
	}
}

func TestHandleError_ValidationWithMessage(t *testing.T) {
	rr := httptest.NewRecorder()
	err := domain.WrapError(domain.ErrValidation, "email is required")

	HandleError(rr, err)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", rr.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if response.Error.Message != "email is required" {
		t.Errorf("Expected message 'email is required', got '%s'", response.Error.Message)
	}
}

func TestHandleErrorCtx_DomainError(t *testing.T) {
	rr := httptest.NewRecorder()
	traceID := "test-trace-123"
	ctx := ctxutil.NewRequestIDContext(context.Background(), traceID)
	err := domainerrors.NewDomain(domainerrors.CodeNotFound, "resource not found")

	HandleErrorCtx(rr, ctx, err)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}

	var response Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if response.Error == nil {
		t.Fatal("Expected error in response")
	}
	if response.Error.Code != domainerrors.CodeNotFound {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeNotFound, response.Error.Code)
	}
	if response.Error.Message != "resource not found" {
		t.Errorf("Expected message 'resource not found', got %s", response.Error.Message)
	}
	if response.Meta == nil || response.Meta.TraceID != traceID {
		t.Errorf("Expected trace_id %s, got %v", traceID, response.Meta)
	}
}

func TestHandleErrorCtx_DomainErrorWithHint(t *testing.T) {
	rr := httptest.NewRecorder()
	traceID := "test-trace-456"
	ctx := ctxutil.NewRequestIDContext(context.Background(), traceID)
	err := domainerrors.NewDomainWithHint(
		domainerrors.CodeValidationError,
		"invalid input",
		"check your request format",
	)

	HandleErrorCtx(rr, ctx, err)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", rr.Code)
	}

	var response Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if response.Error == nil {
		t.Fatal("Expected error in response")
	}
	if response.Error.Hint != "check your request format" {
		t.Errorf("Expected hint 'check your request format', got %s", response.Error.Hint)
	}
}

func TestHandleErrorCtx_LegacyError(t *testing.T) {
	rr := httptest.NewRecorder()
	traceID := "test-trace-789"
	ctx := ctxutil.NewRequestIDContext(context.Background(), traceID)
	err := domain.WrapError(domain.ErrForbidden, "access denied")

	HandleErrorCtx(rr, ctx, err)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}

	var response Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if response.Error == nil {
		t.Fatal("Expected error in response")
	}
	if response.Error.Code != CodeForbidden {
		t.Errorf("Expected code %s, got %s", CodeForbidden, response.Error.Code)
	}
	if response.Meta == nil || response.Meta.TraceID != traceID {
		t.Errorf("Expected trace_id %s, got %v", traceID, response.Meta)
	}
}
