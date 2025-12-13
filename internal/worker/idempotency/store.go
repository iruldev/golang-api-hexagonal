package idempotency

import (
	"context"
	"time"
)

// Store is the interface for idempotency key storage.
// Implementations must be thread-safe and handle concurrent access.
type Store interface {
	// Check atomically checks if a key is new and marks it as seen.
	// Returns true if this is the first time seeing the key (new).
	// Returns false if the key exists (duplicate).
	// The key will expire after the specified TTL.
	Check(ctx context.Context, key string, ttl time.Duration) (isNew bool, err error)

	// StoreResult stores a result for a given idempotency key.
	// This is optional and used for read idempotency (returning cached results).
	StoreResult(ctx context.Context, key string, result []byte, ttl time.Duration) error

	// GetResult retrieves a stored result for a given idempotency key.
	// Returns (result, true, nil) if found.
	// Returns (nil, false, nil) if not found.
	// Returns (nil, false, error) on Redis error.
	GetResult(ctx context.Context, key string) ([]byte, bool, error)
}
