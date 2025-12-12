// Package note provides the Note usecase implementation.
// This is an example usecase demonstrating hexagonal architecture patterns.
package note

import (
	"context"

	"github.com/google/uuid"
	"github.com/iruldev/golang-api-hexagonal/internal/domain/note"
)

// Usecase implements business logic for Note operations.
// It depends on the Repository interface, not a concrete implementation.
type Usecase struct {
	repo note.Repository
}

// NewUsecase creates a new Note usecase.
func NewUsecase(repo note.Repository) *Usecase {
	return &Usecase{repo: repo}
}

// Create creates a new note with validation.
// Returns domain errors for invalid data.
func (u *Usecase) Create(ctx context.Context, title, content string) (*note.Note, error) {
	n := note.NewNote(title, content)

	// Validate before persistence
	if err := n.Validate(); err != nil {
		return nil, err
	}

	if err := u.repo.Create(ctx, n); err != nil {
		return nil, err
	}

	return n, nil
}

// Get retrieves a note by ID.
// Returns ErrNoteNotFound if the note doesn't exist.
func (u *Usecase) Get(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return u.repo.Get(ctx, id)
}

// List retrieves notes with pagination.
// Returns notes, total count, and any error.
func (u *Usecase) List(ctx context.Context, page, pageSize int) ([]*note.Note, int64, error) {
	// Calculate offset from page number
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	return u.repo.List(ctx, pageSize, offset)
}

// Update updates an existing note with validation.
// Returns domain errors for invalid data.
func (u *Usecase) Update(ctx context.Context, id uuid.UUID, title, content string) (*note.Note, error) {
	// Get existing note
	n, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	n.Update(title, content)

	// Validate before persistence
	if err := n.Validate(); err != nil {
		return nil, err
	}

	if err := u.repo.Update(ctx, n); err != nil {
		return nil, err
	}

	return n, nil
}

// Delete removes a note by ID.
// Returns ErrNoteNotFound if the note doesn't exist.
func (u *Usecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}
