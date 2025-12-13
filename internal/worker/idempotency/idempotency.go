// Package idempotency provides idempotency key pattern for async job deduplication.
//
// Idempotency ensures that duplicate jobs don't cause data corruption by tracking
// which jobs have already been processed. Unlike asynq's built-in Unique option
// (which only prevents duplicate enqueuing), this package provides handler-level
// idempotency that protects against duplicate processing even during retries.
//
// # Key Features
//
//   - Handler-level idempotency (protects against duplicate processing)
//   - Configurable TTL (deduplication window)
//   - Fail-open or fail-closed modes for Redis errors
//   - Custom key extraction logic
//   - Optional result caching
//
// # When to Use
//
// Use this package when:
//   - Handler-level idempotency is needed (task re-delivered via retry)
//   - Custom key logic is required (e.g., orderID + productID)
//   - Result caching for duplicates is desired
//   - Control over fail mode (fail-open vs fail-closed) is needed
//
// Use asynq.Unique() when:
//   - Only enqueue-level deduplication is needed
//   - Simple task uniqueness based on task type + payload
//
// For maximum protection, combine both approaches.
//
// # Example Usage
//
// Basic idempotent handler:
//
//	store := idempotency.NewRedisStore(redisClient, "idempotency:")
//	handler := idempotency.IdempotentHandler(
//	    store,
//	    extractKey,
//	    24*time.Hour,
//	    originalHandler,
//	)
//	srv.HandleFunc("order:create", handler)
//
// With fail-closed mode:
//
//	store := idempotency.NewRedisStore(redisClient, "idempotency:",
//	    idempotency.WithFailMode(idempotency.FailClosed),
//	)
package idempotency

import "time"

// DefaultTTL is the default time-to-live for idempotency keys (24 hours).
const DefaultTTL = 24 * time.Hour

// DefaultKeyPrefix is the default Redis key prefix for idempotency keys.
const DefaultKeyPrefix = "idempotency:"

// FailMode determines behavior when Redis is unavailable.
type FailMode int

const (
	// FailOpen processes the task when Redis is unavailable (safe default).
	// Use this for non-critical tasks where processing duplicates is acceptable.
	FailOpen FailMode = iota

	// FailClosed returns an error when Redis is unavailable.
	// Use this for critical tasks where duplicate processing must be prevented.
	FailClosed
)

// String returns the string representation of FailMode.
func (f FailMode) String() string {
	switch f {
	case FailOpen:
		return "fail-open"
	case FailClosed:
		return "fail-closed"
	default:
		return "unknown"
	}
}
