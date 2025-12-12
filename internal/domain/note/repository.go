package note

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the persistence operations for Note entities.
// This interface follows the hexagonal architecture pattern, allowing
// the domain to define what it needs without knowing the implementation.
//
// Implementations should:
//   - Return ErrNoteNotFound when a note doesn't exist
//   - Handle context cancellation appropriately
//   - Use transactions when needed (via TxManager)
type Repository interface {
	// Create persists a new note.
	Create(ctx context.Context, note *Note) error

	// Get retrieves a note by ID.
	// Returns ErrNoteNotFound if the note doesn't exist.
	Get(ctx context.Context, id uuid.UUID) (*Note, error)

	// List retrieves notes with pagination.
	// Results are ordered by created_at DESC.
	List(ctx context.Context, limit, offset int) ([]*Note, int64, error)

	// Update persists changes to an existing note.
	// Returns ErrNoteNotFound if the note doesn't exist.
	Update(ctx context.Context, note *Note) error

	// Delete removes a note by ID.
	// Returns ErrNoteNotFound if the note doesn't exist.
	Delete(ctx context.Context, id uuid.UUID) error
}
