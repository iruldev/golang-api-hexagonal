// Package containers provides testcontainers helpers for integration tests.
//
// This package contains helpers for spinning up Docker containers
// and managing test isolation for integration testing:
//
//   - NewPostgres: create PostgreSQL container
//   - Migrate: run goose migrations
//   - WithTx: run test in transaction with rollback
//   - Truncate: truncate tables for test isolation
//
// See README.md for usage examples.
package containers
