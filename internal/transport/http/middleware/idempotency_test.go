package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

// mockIdempotencyStore is a mock implementation of IdempotencyStore for testing.
type mockIdempotencyStore struct {
	records  map[string]*IdempotencyRecord
	getErr   error
	storeErr error
}

func newMockIdempotencyStore() *mockIdempotencyStore {
	return &mockIdempotencyStore{
		records: make(map[string]*IdempotencyRecord),
	}
}

func (m *mockIdempotencyStore) Get(_ context.Context, key string) (*IdempotencyRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.records[key], nil
}

func (m *mockIdempotencyStore) Store(_ context.Context, record *IdempotencyRecord) error {
	if m.storeErr != nil {
		return m.storeErr
	}
	m.records[record.Key] = record
	return nil
}

func TestIdempotency(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		key            string
		body           string
		existingRecord *IdempotencyRecord
		storeErr       error
		getErr         error
		wantStatus     int
		wantHeader     string // Idempotency-Status value
		wantCode       string // Error code for errors
		wantStored     bool   // Whether record should be stored
	}{
		{
			name:       "POST without key passes through",
			method:     http.MethodPost,
			key:        "",
			body:       `{"email": "test@example.com"}`,
			wantStatus: http.StatusCreated,
			wantStored: false,
		},
		{
			name:       "GET request passes through",
			method:     http.MethodGet,
			key:        "550e8400-e29b-41d4-a716-446655440000",
			wantStatus: http.StatusOK,
			wantStored: false,
		},
		{
			name:       "valid key - first request - stored",
			method:     http.MethodPost,
			key:        "550e8400-e29b-41d4-a716-446655440000",
			body:       `{"email": "test@example.com"}`,
			wantStatus: http.StatusCreated,
			wantHeader: IdempotencyStatusStored,
			wantStored: true,
		},
		{
			name:   "duplicate key - same body - replayed",
			method: http.MethodPost,
			key:    "550e8400-e29b-41d4-a716-446655440000",
			body:   `{"email": "test@example.com"}`,
			existingRecord: &IdempotencyRecord{
				Key:             "550e8400-e29b-41d4-a716-446655440000",
				RequestHash:     computeRequestHash([]byte(`{"email": "test@example.com"}`)),
				StatusCode:      http.StatusCreated,
				ResponseHeaders: http.Header{"Content-Type": []string{"application/json"}},
				ResponseBody:    []byte(`{"id": "123"}`),
			},
			wantStatus: http.StatusCreated,
			wantHeader: IdempotencyStatusReplayed,
			wantStored: false, // Already exists
		},
		{
			name:   "duplicate key - different body - conflict",
			method: http.MethodPost,
			key:    "550e8400-e29b-41d4-a716-446655440000",
			body:   `{"email": "other@example.com"}`,
			existingRecord: &IdempotencyRecord{
				Key:         "550e8400-e29b-41d4-a716-446655440000",
				RequestHash: computeRequestHash([]byte(`{"email": "test@example.com"}`)),
			},
			wantStatus: http.StatusConflict,
			wantCode:   contract.CodeValIdempotencyConflict,
			wantStored: false,
		},
		{
			name:       "invalid key format - not UUID",
			method:     http.MethodPost,
			key:        "not-a-uuid",
			body:       `{"email": "test@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   contract.CodeValIdempotencyKeyInvalid,
			wantStored: false,
		},
		{
			name:       "invalid key format - empty UUID",
			method:     http.MethodPost,
			key:        "00000000-0000-0000-0000-000000000000",
			body:       `{"email": "test@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   contract.CodeValIdempotencyKeyInvalid,
			wantStored: false,
		},
		{
			name:       "store error is ignored",
			method:     http.MethodPost,
			key:        "550e8400-e29b-41d4-a716-446655440000",
			body:       `{"email": "test@example.com"}`,
			storeErr:   context.DeadlineExceeded,
			wantStatus: http.StatusCreated,
			wantHeader: IdempotencyStatusStored,
			wantStored: false, // Store failed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock store
			store := newMockIdempotencyStore()
			store.getErr = tt.getErr
			store.storeErr = tt.storeErr
			if tt.existingRecord != nil {
				store.records[tt.existingRecord.Key] = tt.existingRecord
			}

			// Create middleware
			middleware := Idempotency(IdempotencyConfig{
				Store: store,
				TTL:   time.Hour,
			})

			// Create test handler
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if strings.HasPrefix(tt.method, http.MethodPost) {
					w.WriteHeader(http.StatusCreated)
				} else {
					w.WriteHeader(http.StatusOK)
				}
				_, _ = w.Write([]byte(`{"id": "123"}`))
			}))

			// Create request
			var body *bytes.Reader
			if tt.body != "" {
				body = bytes.NewReader([]byte(tt.body))
			} else {
				body = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(tt.method, "/api/v1/users", body)
			if tt.key != "" {
				req.Header.Set(IdempotencyKeyHeader, tt.key)
			}

			// Execute request
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.wantStatus {
				t.Errorf("status code = %d, want %d", rr.Code, tt.wantStatus)
			}

			// Check Idempotency-Status header
			if tt.wantHeader != "" {
				gotHeader := rr.Header().Get(IdempotencyStatusHeader)
				if gotHeader != tt.wantHeader {
					t.Errorf("Idempotency-Status = %q, want %q", gotHeader, tt.wantHeader)
				}
			}

			// Check error code for error responses
			if tt.wantCode != "" {
				var problem struct {
					Code string `json:"code"`
				}
				if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil {
					t.Fatalf("failed to unmarshal error response: %v", err)
				}
				if problem.Code != tt.wantCode {
					t.Errorf("error code = %q, want %q", problem.Code, tt.wantCode)
				}
			}

			// Check if record was stored
			if tt.wantStored {
				if _, ok := store.records[tt.key]; !ok {
					t.Error("expected record to be stored, but it wasn't")
				}
			}
		})
	}
}

func TestIdempotency_ResponseReplay(t *testing.T) {
	// Setup
	store := newMockIdempotencyStore()
	key := "550e8400-e29b-41d4-a716-446655440000"
	body := `{"email": "test@example.com"}`

	// Pre-store a record
	store.records[key] = &IdempotencyRecord{
		Key:         key,
		RequestHash: computeRequestHash([]byte(body)),
		StatusCode:  http.StatusCreated,
		ResponseHeaders: http.Header{
			"Content-Type":    []string{"application/json"},
			"X-Custom-Header": []string{"custom-value"},
		},
		ResponseBody: []byte(`{"id": "stored-123", "cached": true}`),
	}

	// Create middleware
	middleware := Idempotency(IdempotencyConfig{
		Store: store,
	})

	// Handler should NOT be called for replayed requests
	handlerCalled := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusInternalServerError)
	}))

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader([]byte(body)))
	req.Header.Set(IdempotencyKeyHeader, key)

	// Execute
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Verify
	if handlerCalled {
		t.Error("handler was called for replayed request")
	}

	if rr.Code != http.StatusCreated {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusCreated)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", rr.Header().Get("Content-Type"), "application/json")
	}

	if rr.Header().Get("X-Custom-Header") != "custom-value" {
		t.Errorf("X-Custom-Header = %q, want %q", rr.Header().Get("X-Custom-Header"), "custom-value")
	}

	if rr.Header().Get(IdempotencyStatusHeader) != IdempotencyStatusReplayed {
		t.Errorf("Idempotency-Status = %q, want %q", rr.Header().Get(IdempotencyStatusHeader), IdempotencyStatusReplayed)
	}

	expectedBody := `{"id": "stored-123", "cached": true}`
	if rr.Body.String() != expectedBody {
		t.Errorf("body = %q, want %q", rr.Body.String(), expectedBody)
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid UUID v4", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid UUID v4 lowercase", "a1b2c3d4-e5f6-41a7-8b9c-0d1e2f3a4b5c", true},
		{"valid UUID v1", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", true}, // Accepts any valid UUID version
		{"nil UUID", "00000000-0000-0000-0000-000000000000", false},
		{"empty string", "", false},
		{"not a UUID", "not-a-uuid", false},
		{"too short", "550e8400-e29b-41d4", false},
		{"invalid characters", "550e8400-e29b-41d4-zzzz-446655440000", false},
		{"no dashes", "550e8400e29b41d4a716446655440000", true}, // UUID without dashes is valid
		{"uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidUUID(tt.input)
			if got != tt.want {
				t.Errorf("isValidUUID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestComputeRequestHash(t *testing.T) {
	// Same content should produce same hash
	body1 := []byte(`{"email": "test@example.com"}`)
	body2 := []byte(`{"email": "test@example.com"}`)
	body3 := []byte(`{"email": "other@example.com"}`)

	hash1 := computeRequestHash(body1)
	hash2 := computeRequestHash(body2)
	hash3 := computeRequestHash(body3)

	if hash1 != hash2 {
		t.Errorf("same content should produce same hash: %s != %s", hash1, hash2)
	}

	if hash1 == hash3 {
		t.Errorf("different content should produce different hash: %s == %s", hash1, hash3)
	}

	// Hash should be 64 hex characters (SHA-256)
	if len(hash1) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash1))
	}
}

func TestIdempotency_DefaultTTL(t *testing.T) {
	store := newMockIdempotencyStore()
	key := "550e8400-e29b-41d4-a716-446655440000"
	body := `{"email": "test@example.com"}`

	// Create middleware with no TTL (should use default)
	middleware := Idempotency(IdempotencyConfig{
		Store: store,
		// TTL not set - should default to 24 hours
	})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader([]byte(body)))
	req.Header.Set(IdempotencyKeyHeader, key)

	// Execute
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that record was stored with default TTL
	stored := store.records[key]
	if stored == nil {
		t.Fatal("record was not stored")
	}

	expectedTTL := DefaultIdempotencyTTL
	actualTTL := stored.ExpiresAt.Sub(stored.CreatedAt)
	if actualTTL != expectedTTL {
		t.Errorf("TTL = %v, want %v", actualTTL, expectedTTL)
	}
}

func TestIdempotency_GetError(t *testing.T) {
	store := newMockIdempotencyStore()
	store.getErr = context.DeadlineExceeded

	middleware := Idempotency(IdempotencyConfig{
		Store: store,
	})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader([]byte(`{}`)))
	req.Header.Set(IdempotencyKeyHeader, "550e8400-e29b-41d4-a716-446655440000")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	var problem struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if problem.Code != contract.CodeSysInternal {
		t.Errorf("error code = %q, want %q", problem.Code, contract.CodeSysInternal)
	}
}

// TestIdempotency_ErrorResponseContentType verifies that error responses
// have the correct Content-Type header (application/problem+json).
func TestIdempotency_ErrorResponseContentType(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		wantStatus int
	}{
		{
			name:       "invalid key returns problem+json",
			key:        "not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "nil UUID returns problem+json",
			key:        "00000000-0000-0000-0000-000000000000",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockIdempotencyStore()
			middleware := Idempotency(IdempotencyConfig{Store: store})

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
			}))

			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader([]byte(`{}`)))
			req.Header.Set(IdempotencyKeyHeader, tt.key)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code = %d, want %d", rr.Code, tt.wantStatus)
			}

			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/problem+json" {
				t.Errorf("Content-Type = %q, want %q", contentType, "application/problem+json")
			}
		})
	}
}

// errorReader is a reader that always returns an error.
type errorReader struct {
	err error
}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, r.err
}

// TestIdempotency_BodyReadError verifies error handling when request body cannot be read.
func TestIdempotency_BodyReadError(t *testing.T) {
	store := newMockIdempotencyStore()
	middleware := Idempotency(IdempotencyConfig{Store: store})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	// Create request with a failing reader
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", &errorReader{err: context.DeadlineExceeded})
	req.Header.Set(IdempotencyKeyHeader, "550e8400-e29b-41d4-a716-446655440000")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should return 500 Internal Server Error
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	// Verify it's a proper problem+json response
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/problem+json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/problem+json")
	}

	var problem struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if problem.Code != contract.CodeSysInternal {
		t.Errorf("error code = %q, want %q", problem.Code, contract.CodeSysInternal)
	}
}

// TestIdempotency_ConflictResponseContentType verifies 409 Conflict responses
// have correct Content-Type header.
func TestIdempotency_ConflictResponseContentType(t *testing.T) {
	store := newMockIdempotencyStore()
	key := "550e8400-e29b-41d4-a716-446655440000"

	// Pre-store a record with different body hash
	store.records[key] = &IdempotencyRecord{
		Key:         key,
		RequestHash: computeRequestHash([]byte(`{"email": "original@example.com"}`)),
	}

	middleware := Idempotency(IdempotencyConfig{Store: store})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	// Send request with same key but different body
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader([]byte(`{"email": "different@example.com"}`)))
	req.Header.Set(IdempotencyKeyHeader, key)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusConflict)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/problem+json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/problem+json")
	}
}
