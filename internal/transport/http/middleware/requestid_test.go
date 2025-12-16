package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID_GeneratesNewID(t *testing.T) {
	// Arrange
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request ID is in context
		requestID := GetRequestID(r.Context())
		assert.NotEmpty(t, requestID, "requestId should be in context")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	RequestID(handler).ServeHTTP(rec, req)

	// Assert
	responseID := rec.Header().Get(headerXRequestID)
	assert.NotEmpty(t, responseID, "X-Request-ID should be in response header")
	assert.Len(t, responseID, 32, "request ID should be 32 hex characters")
}

func TestRequestID_PassthroughExistingID(t *testing.T) {
	// Arrange
	providedID := "test-request-id-12345"

	var capturedID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(headerXRequestID, providedID)
	rec := httptest.NewRecorder()

	// Act
	RequestID(handler).ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, providedID, capturedID, "should passthrough provided request ID")
	assert.Equal(t, providedID, rec.Header().Get(headerXRequestID), "response header should contain provided ID")
}

func TestRequestID_ResponseHeader(t *testing.T) {
	// Arrange
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	RequestID(handler).ServeHTTP(rec, req)

	// Assert
	responseID := rec.Header().Get(headerXRequestID)
	require.NotEmpty(t, responseID, "X-Request-ID should be set in response")
}

func TestGetRequestID_ReturnsCorrectValue(t *testing.T) {
	// Arrange
	expectedID := "test-request-id"
	ctx := context.WithValue(context.Background(), requestIDKey, expectedID)

	// Act
	actualID := GetRequestID(ctx)

	// Assert
	assert.Equal(t, expectedID, actualID, "should return correct request ID from context")
}

func TestGetRequestID_ReturnsEmptyForNoValue(t *testing.T) {
	// Act
	actualID := GetRequestID(context.Background())

	// Assert
	assert.Empty(t, actualID, "should return empty string when no request ID in context")
}

func TestGetRequestID_ReturnsEmptyForWrongType(t *testing.T) {
	// Arrange
	ctx := context.WithValue(context.Background(), requestIDKey, 12345) // wrong type

	// Act
	actualID := GetRequestID(ctx)

	// Assert
	assert.Empty(t, actualID, "should return empty string when value is wrong type")
}

func TestGenerateRequestID_HexFormat(t *testing.T) {
	// Act
	id := generateRequestID()

	// Assert
	assert.Len(t, id, 32, "generated ID should be 32 hex characters (16 bytes)")

	// Verify it's valid hex
	for _, c := range id {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"character %c should be valid lowercase hex", c)
	}
}

func TestGenerateRequestID_Unique(t *testing.T) {
	// Generate multiple IDs and verify uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateRequestID()
		assert.False(t, ids[id], "generated IDs should be unique")
		ids[id] = true
	}
}

func TestRequestID_MultipleRequests(t *testing.T) {
	// Arrange
	var capturedIDs []string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedIDs = append(capturedIDs, GetRequestID(r.Context()))
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID(handler)

	// Act
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)
	}

	// Assert
	assert.Len(t, capturedIDs, 3, "should have 3 request IDs")
	// Verify all IDs are unique
	uniqueIDs := make(map[string]bool)
	for _, id := range capturedIDs {
		assert.False(t, uniqueIDs[id], "each request should get a unique ID")
		uniqueIDs[id] = true
	}
}
