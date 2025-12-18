// Package postgres provides PostgreSQL database adapters and repository implementations.
package postgres

import (
	"github.com/google/uuid"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// UUIDGenerator implements domain.IDGenerator using UUID v7.
type UUIDGenerator struct{}

// NewIDGenerator creates a new IDGenerator that generates UUID v7 identifiers.
func NewIDGenerator() domain.IDGenerator {
	return &UUIDGenerator{}
}

// NewID generates a new UUID v7 identifier.
func (g *UUIDGenerator) NewID() domain.ID {
	id, _ := uuid.NewV7()
	return domain.ID(id.String())
}
