// Package postgres provides PostgreSQL adapters and repository factories.
package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"

	notedom "github.com/iruldev/golang-api-hexagonal/internal/domain/note"
	noterepo "github.com/iruldev/golang-api-hexagonal/internal/infra/postgres/note"
)

// NewNoteRepository creates a new PostgreSQL note repository.
// This is a convenience function to create the repository from the main package.
func NewNoteRepository(pool *pgxpool.Pool) notedom.Repository {
	return noterepo.NewRepository(pool)
}
