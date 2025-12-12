package note

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewNote(t *testing.T) {
	// Arrange
	title := "Test Title"
	content := "Test Content"

	// Act
	n := NewNote(title, content)

	// Assert
	if n.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if n.Title != title {
		t.Errorf("expected title %q, got %q", title, n.Title)
	}
	if n.Content != content {
		t.Errorf("expected content %q, got %q", content, n.Content)
	}
	if n.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if n.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestNote_Validate(t *testing.T) {
	tests := []struct {
		name    string
		note    *Note
		wantErr error
	}{
		{
			name: "valid note",
			note: &Note{
				ID:        uuid.New(),
				Title:     "Valid Title",
				Content:   "Valid Content",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "empty title",
			note: &Note{
				ID:        uuid.New(),
				Title:     "",
				Content:   "Content",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: ErrEmptyTitle,
		},
		{
			name: "title too long",
			note: &Note{
				ID:        uuid.New(),
				Title:     strings.Repeat("a", MaxTitleLength+1),
				Content:   "Content",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: ErrTitleTooLong,
		},
		{
			name: "max length title is valid",
			note: &Note{
				ID:        uuid.New(),
				Title:     strings.Repeat("a", MaxTitleLength),
				Content:   "Content",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "empty content is valid",
			note: &Note{
				ID:        uuid.New(),
				Title:     "Title",
				Content:   "",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.note.Validate()

			// Assert
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNote_Update(t *testing.T) {
	// Arrange
	n := NewNote("Original Title", "Original Content")
	originalUpdatedAt := n.UpdatedAt

	// Wait a bit to ensure timestamp changes
	time.Sleep(time.Millisecond)

	// Act
	n.Update("New Title", "New Content")

	// Assert
	if n.Title != "New Title" {
		t.Errorf("expected title %q, got %q", "New Title", n.Title)
	}
	if n.Content != "New Content" {
		t.Errorf("expected content %q, got %q", "New Content", n.Content)
	}
	if !n.UpdatedAt.After(originalUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}
