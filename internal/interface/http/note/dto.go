// Package note provides HTTP handlers for Note operations.
package note

import (
	"time"

	"github.com/google/uuid"
	"github.com/iruldev/golang-api-hexagonal/internal/domain/note"
)

// CreateNoteRequest represents the request body for creating a note.
type CreateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// UpdateNoteRequest represents the request body for updating a note.
type UpdateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// NoteResponse represents a note in API responses.
type NoteResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListNotesResponse represents the response for listing notes.
type ListNotesResponse struct {
	Notes      []NoteResponse `json:"notes"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// ToNoteResponse converts a domain Note to an API response.
func ToNoteResponse(n *note.Note) NoteResponse {
	return NoteResponse{
		ID:        n.ID.String(),
		Title:     n.Title,
		Content:   n.Content,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

// ToNoteResponses converts a slice of domain Notes to API responses.
func ToNoteResponses(notes []*note.Note) []NoteResponse {
	result := make([]NoteResponse, len(notes))
	for i, n := range notes {
		result[i] = ToNoteResponse(n)
	}
	return result
}

// ParseUUID parses a string to a UUID.
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
