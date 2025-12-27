// Package containers provides testcontainers helpers for integration tests.
//
// This package contains helpers for spinning up Docker containers
// for integration testing:
//
//   - NewPostgres: create PostgreSQL container (implemented)
//
// Planned helpers (to be implemented in Story 2.2):
//
//   - Migrate: run goose migrations
//   - WithTx: run test in transaction with rollback
//   - Truncate: truncate tables for test isolation
package containers
