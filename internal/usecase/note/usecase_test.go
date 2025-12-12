package note

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/iruldev/golang-api-hexagonal/internal/domain/note"
)

// MockRepository is a mock implementation of note.Repository.
type MockRepository struct {
	CreateFunc func(ctx context.Context, n *note.Note) error
	GetFunc    func(ctx context.Context, id uuid.UUID) (*note.Note, error)
	ListFunc   func(ctx context.Context, limit, offset int) ([]*note.Note, int64, error)
	UpdateFunc func(ctx context.Context, n *note.Note) error
	DeleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *MockRepository) Create(ctx context.Context, n *note.Note) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, n)
	}
	return nil
}

func (m *MockRepository) Get(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockRepository) Update(ctx context.Context, n *note.Note) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, n)
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func TestUsecase_Create(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		content    string
		repoErr    error
		wantErr    error
		expectSave bool
	}{
		{
			name:       "success",
			title:      "Valid Title",
			content:    "Valid Content",
			repoErr:    nil,
			wantErr:    nil,
			expectSave: true,
		},
		{
			name:       "empty title validation error",
			title:      "",
			content:    "Content",
			repoErr:    nil,
			wantErr:    note.ErrEmptyTitle,
			expectSave: false,
		},
		{
			name:       "repository error",
			title:      "Title",
			content:    "Content",
			repoErr:    errors.New("db error"),
			wantErr:    errors.New("db error"),
			expectSave: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			saved := false
			repo := &MockRepository{
				CreateFunc: func(ctx context.Context, n *note.Note) error {
					saved = true
					return tt.repoErr
				},
			}
			uc := NewUsecase(repo)

			// Act
			result, err := uc.Create(context.Background(), tt.title, tt.content)

			// Assert
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected non-nil result")
				} else if result.Title != tt.title {
					t.Errorf("expected title %q, got %q", tt.title, result.Title)
				}
			}
			if saved != tt.expectSave {
				t.Errorf("expected save=%v, got %v", tt.expectSave, saved)
			}
		})
	}
}

func TestUsecase_Get(t *testing.T) {
	tests := []struct {
		name    string
		id      uuid.UUID
		note    *note.Note
		repoErr error
		wantErr error
	}{
		{
			name:    "success",
			id:      uuid.New(),
			note:    note.NewNote("Title", "Content"),
			wantErr: nil,
		},
		{
			name:    "not found",
			id:      uuid.New(),
			note:    nil,
			repoErr: note.ErrNoteNotFound,
			wantErr: note.ErrNoteNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &MockRepository{
				GetFunc: func(ctx context.Context, id uuid.UUID) (*note.Note, error) {
					return tt.note, tt.repoErr
				},
			}
			uc := NewUsecase(repo)

			// Act
			result, err := uc.Get(context.Background(), tt.id)

			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestUsecase_List(t *testing.T) {
	// Arrange
	notes := []*note.Note{
		note.NewNote("Title 1", "Content 1"),
		note.NewNote("Title 2", "Content 2"),
	}
	repo := &MockRepository{
		ListFunc: func(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
			return notes, 2, nil
		},
	}
	uc := NewUsecase(repo)

	// Act
	result, total, err := uc.List(context.Background(), 1, 10)

	// Assert
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 notes, got %d", len(result))
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
}

func TestUsecase_Update(t *testing.T) {
	existingNote := note.NewNote("Original", "Content")

	tests := []struct {
		name    string
		id      uuid.UUID
		title   string
		content string
		getErr  error
		updErr  error
		wantErr error
	}{
		{
			name:    "success",
			id:      existingNote.ID,
			title:   "Updated Title",
			content: "Updated Content",
			wantErr: nil,
		},
		{
			name:    "not found",
			id:      uuid.New(),
			title:   "Title",
			content: "Content",
			getErr:  note.ErrNoteNotFound,
			wantErr: note.ErrNoteNotFound,
		},
		{
			name:    "validation error",
			id:      existingNote.ID,
			title:   "",
			content: "Content",
			wantErr: note.ErrEmptyTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &MockRepository{
				GetFunc: func(ctx context.Context, id uuid.UUID) (*note.Note, error) {
					if tt.getErr != nil {
						return nil, tt.getErr
					}
					return note.NewNote("Original", "Content"), nil
				},
				UpdateFunc: func(ctx context.Context, n *note.Note) error {
					return tt.updErr
				},
			}
			uc := NewUsecase(repo)

			// Act
			result, err := uc.Update(context.Background(), tt.id, tt.title, tt.content)

			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestUsecase_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uuid.UUID
		repoErr error
		wantErr error
	}{
		{
			name:    "success",
			id:      uuid.New(),
			repoErr: nil,
			wantErr: nil,
		},
		{
			name:    "not found",
			id:      uuid.New(),
			repoErr: note.ErrNoteNotFound,
			wantErr: note.ErrNoteNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &MockRepository{
				DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
					return tt.repoErr
				},
			}
			uc := NewUsecase(repo)

			// Act
			err := uc.Delete(context.Background(), tt.id)

			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
