package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// envelopeResponse wraps the response for testing
type envelopeResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

func TestExampleHandler_ReturnsJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/example", nil)
	rr := httptest.NewRecorder()

	ExampleHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestExampleHandler_EnvelopeFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/example", nil)
	rr := httptest.NewRecorder()

	ExampleHandler(rr, req)

	var envelope response.SuccessResponse
	if err := json.NewDecoder(rr.Body).Decode(&envelope); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !envelope.Success {
		t.Error("Expected success to be true")
	}

	if envelope.Data == nil {
		t.Error("Expected data to be present")
	}
}

func TestExampleHandler_DataContent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/example", nil)
	rr := httptest.NewRecorder()

	ExampleHandler(rr, req)

	var envelope envelopeResponse
	if err := json.NewDecoder(rr.Body).Decode(&envelope); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	var data ExampleData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	if data.Message != "Example handler working correctly" {
		t.Errorf("Unexpected message: %s", data.Message)
	}
}

func TestHealthHandler_ReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestHealthHandler_EnvelopeFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	var envelope envelopeResponse
	if err := json.NewDecoder(rr.Body).Decode(&envelope); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !envelope.Success {
		t.Error("Expected success to be true")
	}

	var data HealthData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	if data.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", data.Status)
	}
}
