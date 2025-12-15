// Package wrapper provides context-aware wrapper functions for database,
// HTTP, and Redis operations.
//
// This package enforces consistent context propagation across I/O operations
// by providing wrapper functions that:
//   - Require context as the first parameter
//   - Apply default timeouts when context has no deadline
//   - Return early if context is already done
//   - Preserve existing deadlines (never overwrite)
//
// Default timeouts:
//   - Database operations: 30 seconds
//   - HTTP requests: 30 seconds
//   - Redis operations: 30 seconds
//
// Usage:
//
//	// Database query with automatic timeout
//	rows, err := wrapper.Query(ctx, pool, "SELECT * FROM users")
//
//	// HTTP request with automatic timeout
//	resp, err := wrapper.DoRequest(ctx, client, req)
//
//	// Redis operation with context check
//	err := wrapper.DoRedis(ctx, func(ctx context.Context) error { return rdb.Set(ctx, ...) })
//
// This package is part of the infrastructure layer and can only import
// from the domain layer, following hexagonal architecture principles.
package wrapper
