package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// envelopeResponse wraps the response for testing with new Envelope format
type envelopeResponse struct {
	Data json.RawMessage `json:"data"`
	Meta *response.Meta  `json:"meta"`
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

	var envelope envelopeResponse
	if err := json.NewDecoder(rr.Body).Decode(&envelope); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if envelope.Meta == nil {
		t.Error("Expected meta to be present")
	}

	if envelope.Data == nil {
		t.Error("Expected data to be present")
	}
}

func TestExampleHandler_DataContent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/example", nil)
	rr := httptest.NewRecorder()

	ExampleHandler(rr, req)

	// ExampleHandler now uses new format
	var envelope envelopeResponse
	if err := json.NewDecoder(rr.Body).Decode(&envelope); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	var exampleData ExampleData
	if err := json.Unmarshal(data, &exampleData); err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	if exampleData.Message != "Example handler working correctly" {
		t.Errorf("Unexpected message: %s", exampleData.Message)
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

	// Check data is present
	if envelope.Data == nil {
		t.Error("Expected data to be present")
	}

	// Check meta is present with trace_id
	if envelope.Meta == nil {
		t.Fatal("Expected meta to be present")
	}
	if envelope.Meta.TraceID == "" {
		t.Error("Expected meta.trace_id to be present")
	}

	var data HealthData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	if data.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", data.Status)
	}
}
