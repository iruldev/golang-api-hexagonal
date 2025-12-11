// Package runtimeutil provides runtime utility interfaces for external services.
// These interfaces allow for swappable implementations (e.g., Redis → Memcached).
package runtimeutil

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss indicates the key was not found in cache.
var ErrCacheMiss = errors.New("cache: key not found")

// Cache defines caching abstraction for swappable implementations.
// Implement this interface for Redis, Memcached, or in-memory cache.
//
// Usage Example:
//
//	// Get a value
//	value, err := cache.Get(ctx, "user:123")
//	if errors.Is(err, runtimeutil.ErrCacheMiss) {
//	    // Key not found, fetch from DB and set cache
//	    value = fetchFromDB()
//	    cache.Set(ctx, "user:123", value, 5*time.Minute)
//	}
//
//	// Delete on update
//	cache.Delete(ctx, "user:123")
//
// Implementing Redis Adapter:
//
//	type RedisCache struct {
//	    client *redis.Client
//	}
//
//	func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
//	    val, err := c.client.Get(ctx, key).Bytes()
//	    if err == redis.Nil {
//	        return nil, runtimeutil.ErrCacheMiss
//	    }
//	    return val, err
//	}
type Cache interface {
	// Get retrieves a value by key.
	// Returns ErrCacheMiss if the key does not exist.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with optional TTL.
	// Zero TTL means no expiration.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value by key.
	// Returns nil if key does not exist.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)
}

// NopCache is a no-op cache implementation for testing.
type NopCache struct{}

// NewNopCache creates a new NopCache.
func NewNopCache() Cache {
	return &NopCache{}
}

func (c *NopCache) Get(_ context.Context, _ string) ([]byte, error) {
	return nil, ErrCacheMiss
}

func (c *NopCache) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}

func (c *NopCache) Delete(_ context.Context, _ string) error {
	return nil
}

func (c *NopCache) Exists(_ context.Context, _ string) (bool, error) {
	return false, nil
}
