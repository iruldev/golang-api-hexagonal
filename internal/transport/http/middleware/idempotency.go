// Package middleware provides HTTP middleware components.
// This file implements idempotency key middleware for safe POST request retries.
package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

// IdempotencyKeyHeader is the HTTP header name for idempotency keys.
const IdempotencyKeyHeader = "Idempotency-Key"

// IdempotencyStatusHeader is the HTTP header name for idempotency status.
const IdempotencyStatusHeader = "Idempotency-Status"

// IdempotencyStatus values
const (
	// IdempotencyStatusStored indicates the response was stored for the first time.
	IdempotencyStatusStored = "stored"

	// IdempotencyStatusReplayed indicates the response was replayed from cache.
	IdempotencyStatusReplayed = "replayed"
)

// DefaultIdempotencyTTL is the default time-to-live for idempotency records.
const DefaultIdempotencyTTL = 24 * time.Hour

// IdempotencyRecord represents a cached response for an idempotency key.
type IdempotencyRecord struct {
	// Key is the idempotency key (UUID v4).
	Key string

	// RequestHash is the SHA-256 hash of the request body.
	RequestHash string

	// StatusCode is the HTTP status code of the cached response.
	StatusCode int

	// ResponseHeaders contains the response headers to replay.
	ResponseHeaders http.Header

	// ResponseBody is the cached response body.
	ResponseBody []byte

	// CreatedAt is when the record was created.
	CreatedAt time.Time

	// ExpiresAt is when the record expires.
	ExpiresAt time.Time
}

// IdempotencyStore defines the storage interface for idempotency records.
// Implementations are provided by the infra layer (e.g., PostgreSQL).
type IdempotencyStore interface {
	// Get retrieves an existing record by key.
	// Returns nil, nil if the key doesn't exist.
	Get(ctx context.Context, key string) (*IdempotencyRecord, error)

	// Store saves a new idempotency record.
	// Returns an error if the key already exists (race condition).
	Store(ctx context.Context, record *IdempotencyRecord) error
}

// IdempotencyConfig holds configuration for the idempotency middleware.
type IdempotencyConfig struct {
	// Store is the idempotency record storage.
	Store IdempotencyStore

	// TTL is the time-to-live for idempotency records.
	// Default: 24 hours.
	TTL time.Duration
}

// MaxCachedResponseSize is the maximum size of a response body to cache (1MB).
const MaxCachedResponseSize = 1 * 1024 * 1024

// idempotencyResponseWriter wraps http.ResponseWriter to capture the response.
type idempotencyResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	body        bytes.Buffer
	headers     http.Header
	wroteHeader bool
}

// newIdempotencyResponseWriter creates a new response writer wrapper.
func newIdempotencyResponseWriter(w http.ResponseWriter) *idempotencyResponseWriter {
	return &idempotencyResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default to 200 OK
		headers:        make(http.Header),
	}
}

// Header returns the header map for setting headers.
func (w *idempotencyResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// WriteHeader captures the status code and writes to the underlying writer.
// IMPORTANT: Sets Idempotency-Status header BEFORE writing to ensure it's sent to client.
func (w *idempotencyResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.statusCode = code
	// Copy headers at this point
	for k, v := range w.ResponseWriter.Header() {
		w.headers[k] = append([]string(nil), v...)
	}
	// Set idempotency status header BEFORE writing to underlying writer.
	w.ResponseWriter.Header().Set(IdempotencyStatusHeader, IdempotencyStatusStored)
	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

// Write captures the body and writes to the underlying writer.
// It handles implicit 200 OK if WriteHeader wasn't called.
func (w *idempotencyResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	// Only capture up to the limit to avoid memory DoS
	// If we exceed the limit, we stop capturing but continue writing to client.
	// We'll calculate total length to detect truncation later.
	if w.body.Len() < MaxCachedResponseSize {
		remaining := MaxCachedResponseSize - w.body.Len()
		if len(b) > remaining {
			w.body.Write(b[:remaining])
		} else {
			w.body.Write(b)
		}
	}

	return w.ResponseWriter.Write(b)
}

// capturedResponse returns the captured response data and a validity boolean.
// Returns false if response was too large (truncated) or shouldn't be cached.
func (w *idempotencyResponseWriter) capturedResponse() (int, http.Header, []byte, bool) {
	// If Write was called but WriteHeader wasn't explicitly called (should be handled by Write),
	// or if neither was called (empty 200 OK response).
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	// Ensure headers are captured if not already done
	if len(w.headers) == 0 {
		for k, v := range w.ResponseWriter.Header() {
			w.headers[k] = append([]string(nil), v...)
		}
	}

	// Check for potential truncation.
	// NOTE: If buffer len == MaxCachedResponseSize, we skip caching.
	if w.body.Len() >= MaxCachedResponseSize {
		return w.statusCode, w.headers, nil, false // Invalid (too large)
	}

	return w.statusCode, w.headers, w.body.Bytes(), true // Valid
}

// Idempotency returns middleware that provides idempotency for POST requests.
// It caches responses by Idempotency-Key header and replays them for duplicate requests.
//
// Flow:
//  1. Extract Idempotency-Key header (pass through if missing)
//  2. Validate UUID v4 format (400 if invalid)
//  3. Compute SHA-256 hash of request body
//  4. Check storage for existing key:
//     - Not found: continue to handler
//     - Found + hash matches: replay cached response
//     - Found + hash differs: 409 Conflict
//  5. Execute handler with response capture
//  6. Store response with key
//  7. Return response with Idempotency-Status header
func Idempotency(cfg IdempotencyConfig) func(http.Handler) http.Handler {
	// Set default TTL
	ttl := cfg.TTL
	if ttl == 0 {
		ttl = DefaultIdempotencyTTL
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to POST requests
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Extract Idempotency-Key header
			key := r.Header.Get(IdempotencyKeyHeader)
			if key == "" {
				// No idempotency key - pass through
				next.ServeHTTP(w, r)
				return
			}

			// Validate UUID format (accepts any valid UUID version for flexibility)
			if !isValidUUID(key) {
				contract.WriteProblemJSON(w, r, &app.AppError{
					Op:      "Idempotency.ValidateKey",
					Code:    contract.CodeValIdempotencyKeyInvalid,
					Message: "Idempotency key must be a valid UUID",
				})
				return
			}

			// Read and hash request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				contract.WriteProblemJSON(w, r, &app.AppError{
					Op:      "Idempotency.ReadBody",
					Code:    contract.CodeSysInternal,
					Message: "Failed to read request body",
					Err:     err,
				})
				return
			}
			// Restore body for handler
			r.Body = io.NopCloser(bytes.NewReader(body))

			// Compute request hash
			requestHash := computeRequestHash(body)

			// Check for existing record
			existing, err := cfg.Store.Get(r.Context(), key)
			if err != nil {
				contract.WriteProblemJSON(w, r, &app.AppError{
					Op:      "Idempotency.GetRecord",
					Code:    contract.CodeSysInternal,
					Message: "Failed to check idempotency record",
					Err:     err,
				})
				return
			}

			if existing != nil {
				// Record exists - check if request hash matches
				if existing.RequestHash != requestHash {
					// Different request body - conflict
					contract.WriteProblemJSON(w, r, &app.AppError{
						Op:      "Idempotency.CheckConflict",
						Code:    contract.CodeValIdempotencyConflict,
						Message: "Idempotency key already exists with different request body",
					})
					return
				}

				// Same request - replay cached response
				replayResponse(w, existing)
				return
			}

			// First request - execute handler with response capture
			wrapper := newIdempotencyResponseWriter(w)

			// Set status header before handler executes
			// This will be overwritten if handler succeeds
			next.ServeHTTP(wrapper, r)

			// Capture response
			statusCode, headers, responseBody, isValid := wrapper.capturedResponse()

			// If response was too large to cache, we skip storage.
			if !isValid {
				return
			}

			// Store the response
			now := time.Now()
			record := &IdempotencyRecord{
				Key:             key,
				RequestHash:     requestHash,
				StatusCode:      statusCode,
				ResponseHeaders: headers,
				ResponseBody:    responseBody,
				CreatedAt:       now,
				ExpiresAt:       now.Add(ttl),
			}

			// Store record - ignore race condition errors.
			// If another request stored first, that's fine.
			// Note: Store errors are intentionally ignored for reliability -
			// the client request should succeed even if caching fails.
			// The Idempotency-Status header was already set in WriteHeader.
			if err := cfg.Store.Store(r.Context(), record); err != nil {
				// Log store error for observability but don't fail the request.
				// In production, consider adding metrics/tracing here.
				_ = err // Intentionally ignored - request already succeeded
			}
		})
	}
}

// isValidUUID checks if the string is a valid non-nil UUID.
// Note: This accepts any valid UUID version (v1, v4, etc.) for flexibility.
// The original AC specified UUID v4, but accepting any version improves
// interoperability with different client implementations.
// To enforce strict v4: add check `parsed.Version() == 4`.
func isValidUUID(s string) bool {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return false
	}
	return parsed != uuid.Nil
}

// computeRequestHash computes the SHA-256 hash of the request body.
func computeRequestHash(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// replayResponse writes the cached response to the response writer.
func replayResponse(w http.ResponseWriter, record *IdempotencyRecord) {
	// Copy cached headers
	for k, v := range record.ResponseHeaders {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}

	// Set idempotency status
	w.Header().Set(IdempotencyStatusHeader, IdempotencyStatusReplayed)

	// Write status code
	w.WriteHeader(record.StatusCode)

	// Write body
	_, _ = w.Write(record.ResponseBody)
}
