package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

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
	assert.Len(t, responseID, 36, "request ID should be 36 characters (UUID)")
	_, err := uuid.Parse(responseID)
	assert.NoError(t, err, "request ID should be a valid UUID")
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

func TestRequestID_IgnoresTooLongID(t *testing.T) {
	// Arrange
	// 65 characters
	longID := "12345678901234567890123456789012345678901234567890123456789012345"

	var capturedID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(headerXRequestID, longID)
	rec := httptest.NewRecorder()

	// Act
	RequestID(handler).ServeHTTP(rec, req)

	// Assert
	assert.NotEqual(t, longID, capturedID, "should NOT passthrough long request ID")
	assert.Len(t, capturedID, 36, "should generate new valid UUID v7")
	assert.NoError(t, uuid.Validate(capturedID), "should be valid UUID")
}

func TestRequestID_IgnoresInvalidCharset(t *testing.T) {
	// Arrange
	invalidID := "user-id-with-bad-char-$"

	var capturedID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(headerXRequestID, invalidID)
	rec := httptest.NewRecorder()

	// Act
	RequestID(handler).ServeHTTP(rec, req)

	// Assert
	assert.NotEqual(t, invalidID, capturedID, "should NOT passthrough invalid charset")
	assert.Len(t, capturedID, 36, "should generate new valid UUID v7")
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

func TestGenerateRequestID_UUIDFormat(t *testing.T) {
	// Act
	id := generateRequestID()

	// Assert
	assert.Len(t, id, 36, "generated ID should be 36 characters (UUID)")

	// Verify it's valid UUID
	parsed, err := uuid.Parse(id)
	assert.NoError(t, err, "should be valid UUID")
	assert.Equal(t, uuid.Version(7), parsed.Version(), "should be UUID v7")
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
