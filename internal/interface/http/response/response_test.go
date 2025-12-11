package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccess_ReturnsCorrectFormat(t *testing.T) {
	rr := httptest.NewRecorder()

	data := map[string]string{"key": "value"}
	Success(rr, data)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response SuccessResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Data == nil {
		t.Error("Expected data to be present")
	}
}

func TestError_ReturnsCorrectFormat(t *testing.T) {
	rr := httptest.NewRecorder()

	Error(rr, http.StatusBadRequest, ErrBadRequest, "Invalid input")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected success to be false")
	}

	if response.Error.Code != ErrBadRequest {
		t.Errorf("Expected error code %s, got %s", ErrBadRequest, response.Error.Code)
	}

	if response.Error.Message != "Invalid input" {
		t.Errorf("Expected message 'Invalid input', got '%s'", response.Error.Message)
	}
}

func TestSuccessWithStatus_Returns201(t *testing.T) {
	rr := httptest.NewRecorder()

	data := map[string]int{"id": 123}
	SuccessWithStatus(rr, http.StatusCreated, data)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	var response SuccessResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
}

func TestBadRequest_Returns400(t *testing.T) {
	rr := httptest.NewRecorder()

	BadRequest(rr, "Missing required field")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Code != ErrBadRequest {
		t.Errorf("Expected error code %s, got %s", ErrBadRequest, response.Error.Code)
	}
}

func TestNotFound_Returns404(t *testing.T) {
	rr := httptest.NewRecorder()

	NotFound(rr, "User not found")

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Code != ErrNotFound {
		t.Errorf("Expected error code %s, got %s", ErrNotFound, response.Error.Code)
	}
}

func TestUnauthorized_Returns401(t *testing.T) {
	rr := httptest.NewRecorder()

	Unauthorized(rr, "Invalid token")

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestForbidden_Returns403(t *testing.T) {
	rr := httptest.NewRecorder()

	Forbidden(rr, "Insufficient permissions")

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

func TestConflict_Returns409(t *testing.T) {
	rr := httptest.NewRecorder()

	Conflict(rr, "Resource already exists")

	if rr.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", rr.Code)
	}
}

func TestValidationError_Returns422(t *testing.T) {
	rr := httptest.NewRecorder()

	ValidationError(rr, "Email format invalid")

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", rr.Code)
	}
}

func TestInternalServerError_Returns500(t *testing.T) {
	rr := httptest.NewRecorder()

	InternalServerError(rr, "Database connection failed")

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestJSON_SetsContentType(t *testing.T) {
	rr := httptest.NewRecorder()

	JSON(rr, http.StatusOK, "test")

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}
