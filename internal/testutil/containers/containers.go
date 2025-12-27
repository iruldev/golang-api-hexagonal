// Package containers provides testcontainers helpers for integration tests.
//
// This package will contain helpers for spinning up Docker containers
// for integration testing, particularly:
//   - PostgreSQL containers with migrations
//   - Container wait strategies
//   - Transaction isolation helpers
//
// Planned helpers (to be implemented in Story 2.1-2.2):
//   - NewPostgres: create PostgreSQL container
//   - Migrate: run goose migrations
//   - WithTx: run test in transaction with rollback
//   - Truncate: truncate tables for test isolation
package containers
