// Package note provides the Note domain entity and related business logic.
// This is an example domain module demonstrating hexagonal architecture patterns.
package note

import (
	"time"

	"github.com/google/uuid"
)

// Note represents a note in the system.
// This is the core domain entity used across all layers.
type Note struct {
	// ID is the unique identifier for the note.
	ID uuid.UUID `json:"id" db:"id"`

	// Title is the note title (required, max 255 chars).
	Title string `json:"title" db:"title"`

	// Content is the note body (optional).
	Content string `json:"content" db:"content"`

	// CreatedAt is when the note was created.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt is when the note was last updated.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewNote creates a new Note with generated ID and timestamps.
func NewNote(title, content string) *Note {
	now := time.Now().UTC()
	return &Note{
		ID:        uuid.New(),
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the Note entity.
// Returns domain-specific errors for invalid data.
func (n *Note) Validate() error {
	if n.Title == "" {
		return ErrEmptyTitle
	}
	if len(n.Title) > MaxTitleLength {
		return ErrTitleTooLong
	}
	return nil
}

// Update updates the note's title and content.
// Sets UpdatedAt to current time.
func (n *Note) Update(title, content string) {
	n.Title = title
	n.Content = content
	n.UpdatedAt = time.Now().UTC()
}

// MaxTitleLength is the maximum allowed length for note titles.
const MaxTitleLength = 255
