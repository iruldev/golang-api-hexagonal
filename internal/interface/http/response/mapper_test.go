package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

func TestMapError_NotFound(t *testing.T) {
	status, code := MapError(domain.ErrNotFound)

	if status != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", status)
	}
	if code != ErrNotFound {
		t.Errorf("Expected code %s, got %s", ErrNotFound, code)
	}
}

func TestMapError_Validation(t *testing.T) {
	status, code := MapError(domain.ErrValidation)

	if status != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", status)
	}
	if code != ErrValidation {
		t.Errorf("Expected code %s, got %s", ErrValidation, code)
	}
}

func TestMapError_Unauthorized(t *testing.T) {
	status, code := MapError(domain.ErrUnauthorized)

	if status != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", status)
	}
	if code != ErrUnauthorized {
		t.Errorf("Expected code %s, got %s", ErrUnauthorized, code)
	}
}

func TestMapError_Forbidden(t *testing.T) {
	status, code := MapError(domain.ErrForbidden)

	if status != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", status)
	}
	if code != ErrForbidden {
		t.Errorf("Expected code %s, got %s", ErrForbidden, code)
	}
}

func TestMapError_Conflict(t *testing.T) {
	status, code := MapError(domain.ErrConflict)

	if status != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", status)
	}
	if code != ErrConflict {
		t.Errorf("Expected code %s, got %s", ErrConflict, code)
	}
}

func TestMapError_Internal(t *testing.T) {
	status, code := MapError(domain.ErrInternal)

	if status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", status)
	}
	if code != ErrInternalServer {
		t.Errorf("Expected code %s, got %s", ErrInternalServer, code)
	}
}

func TestMapError_Unknown(t *testing.T) {
	// Use an error that doesn't match any domain error
	unknownErr := errors.New("database connection timeout")
	status, code := MapError(unknownErr)

	if status != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for unknown error, got %d", status)
	}
	if code != ErrInternalServer {
		t.Errorf("Expected code %s, got %s", ErrInternalServer, code)
	}
}

func TestMapError_WrappedNotFound(t *testing.T) {
	wrapped := domain.WrapError(domain.ErrNotFound, "user not found")
	status, code := MapError(wrapped)

	if status != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", status)
	}
	if code != ErrNotFound {
		t.Errorf("Expected code %s, got %s", ErrNotFound, code)
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
	if response.Error.Code != ErrNotFound {
		t.Errorf("Expected code %s, got %s", ErrNotFound, response.Error.Code)
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
