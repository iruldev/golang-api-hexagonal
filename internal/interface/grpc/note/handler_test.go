package note

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	notedom "github.com/iruldev/golang-api-hexagonal/internal/domain/note"
	noteuc "github.com/iruldev/golang-api-hexagonal/internal/usecase/note"
	notev1 "github.com/iruldev/golang-api-hexagonal/proto/note/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockRepository implements note.Repository for testing.
type MockRepository struct {
	CreateFunc func(ctx context.Context, note *notedom.Note) error
	GetFunc    func(ctx context.Context, id uuid.UUID) (*notedom.Note, error)
	ListFunc   func(ctx context.Context, limit, offset int) ([]*notedom.Note, int64, error)
	UpdateFunc func(ctx context.Context, note *notedom.Note) error
	DeleteFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *MockRepository) Create(ctx context.Context, note *notedom.Note) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, note)
	}
	return nil
}

func (m *MockRepository) Get(ctx context.Context, id uuid.UUID) (*notedom.Note, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, notedom.ErrNoteNotFound
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*notedom.Note, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockRepository) Update(ctx context.Context, note *notedom.Note) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, note)
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func TestHandler_CreateNote(t *testing.T) {
	tests := []struct {
		name     string
		req      *notev1.CreateNoteRequest
		mockRepo *MockRepository
		wantCode codes.Code
		wantNote bool
	}{
		{
			name: "valid note",
			req: &notev1.CreateNoteRequest{
				Title:   "Test Note",
				Content: "Test content",
			},
			mockRepo: &MockRepository{
				CreateFunc: func(ctx context.Context, note *notedom.Note) error {
					note.ID = uuid.New()
					note.CreatedAt = time.Now()
					note.UpdatedAt = time.Now()
					return nil
				},
			},
			wantCode: codes.OK,
			wantNote: true,
		},
		{
			name: "empty title",
			req: &notev1.CreateNoteRequest{
				Title:   "",
				Content: "Content",
			},
			mockRepo: &MockRepository{},
			wantCode: codes.InvalidArgument,
			wantNote: false,
		},
		{
			name: "title too long",
			req: &notev1.CreateNoteRequest{
				Title:   string(make([]byte, 300)), // 300 chars > MaxTitleLength
				Content: "Content",
			},
			mockRepo: &MockRepository{},
			wantCode: codes.InvalidArgument,
			wantNote: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			uc := noteuc.NewUsecase(tt.mockRepo)
			handler := NewHandler(uc)

			// Act
			resp, err := handler.CreateNote(context.Background(), tt.req)

			// Assert
			if tt.wantCode == codes.OK {
				if err != nil {
					t.Errorf("CreateNote() error = %v, want nil", err)
				}
				if !tt.wantNote || resp == nil || resp.Note == nil {
					t.Errorf("CreateNote() response note missing")
				}
				if resp != nil && resp.Note != nil {
					if resp.Note.Title != tt.req.Title {
						t.Errorf("CreateNote() title = %v, want %v", resp.Note.Title, tt.req.Title)
					}
				}
			} else {
				if err == nil {
					t.Errorf("CreateNote() expected error, got nil")
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("CreateNote() error is not a gRPC status")
				}
				if st.Code() != tt.wantCode {
					t.Errorf("CreateNote() code = %v, want %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

func TestHandler_GetNote(t *testing.T) {
	validID := uuid.New()
	now := time.Now()

	tests := []struct {
		name     string
		req      *notev1.GetNoteRequest
		mockRepo *MockRepository
		wantCode codes.Code
	}{
		{
			name: "valid note found",
			req:  &notev1.GetNoteRequest{Id: validID.String()},
			mockRepo: &MockRepository{
				GetFunc: func(ctx context.Context, id uuid.UUID) (*notedom.Note, error) {
					return &notedom.Note{
						ID:        validID,
						Title:     "Test",
						Content:   "Content",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil
				},
			},
			wantCode: codes.OK,
		},
		{
			name: "note not found",
			req:  &notev1.GetNoteRequest{Id: uuid.New().String()},
			mockRepo: &MockRepository{
				GetFunc: func(ctx context.Context, id uuid.UUID) (*notedom.Note, error) {
					return nil, notedom.ErrNoteNotFound
				},
			},
			wantCode: codes.NotFound,
		},
		{
			name:     "invalid uuid",
			req:      &notev1.GetNoteRequest{Id: "invalid-uuid"},
			mockRepo: &MockRepository{},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noteuc.NewUsecase(tt.mockRepo)
			handler := NewHandler(uc)

			resp, err := handler.GetNote(context.Background(), tt.req)

			if tt.wantCode == codes.OK {
				if err != nil {
					t.Errorf("GetNote() error = %v, want nil", err)
				}
				if resp == nil || resp.Note == nil {
					t.Errorf("GetNote() response note missing")
				}
			} else {
				if err == nil {
					t.Errorf("GetNote() expected error, got nil")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantCode {
					t.Errorf("GetNote() code = %v, want %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

func TestHandler_ListNotes(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		req       *notev1.ListNotesRequest
		mockRepo  *MockRepository
		wantCode  codes.Code
		wantCount int
	}{
		{
			name: "list with results",
			req:  &notev1.ListNotesRequest{PageSize: 10},
			mockRepo: &MockRepository{
				ListFunc: func(ctx context.Context, limit, offset int) ([]*notedom.Note, int64, error) {
					return []*notedom.Note{
						{ID: uuid.New(), Title: "Note 1", CreatedAt: now, UpdatedAt: now},
						{ID: uuid.New(), Title: "Note 2", CreatedAt: now, UpdatedAt: now},
					}, 2, nil
				},
			},
			wantCode:  codes.OK,
			wantCount: 2,
		},
		{
			name: "empty list",
			req:  &notev1.ListNotesRequest{PageSize: 10},
			mockRepo: &MockRepository{
				ListFunc: func(ctx context.Context, limit, offset int) ([]*notedom.Note, int64, error) {
					return nil, 0, nil
				},
			},
			wantCode:  codes.OK,
			wantCount: 0,
		},
		{
			name: "default page size",
			req:  &notev1.ListNotesRequest{}, // No page_size set
			mockRepo: &MockRepository{
				ListFunc: func(ctx context.Context, limit, offset int) ([]*notedom.Note, int64, error) {
					if limit != 10 {
						return nil, 0, errors.New("expected default page size 10")
					}
					return nil, 0, nil
				},
			},
			wantCode:  codes.OK,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noteuc.NewUsecase(tt.mockRepo)
			handler := NewHandler(uc)

			resp, err := handler.ListNotes(context.Background(), tt.req)

			if tt.wantCode == codes.OK {
				if err != nil {
					t.Errorf("ListNotes() error = %v, want nil", err)
				}
				if resp == nil {
					t.Errorf("ListNotes() response is nil")
				}
				if resp != nil && len(resp.Notes) != tt.wantCount {
					t.Errorf("ListNotes() count = %v, want %v", len(resp.Notes), tt.wantCount)
				}
			} else {
				if err == nil {
					t.Errorf("ListNotes() expected error, got nil")
				}
			}
		})
	}
}

func TestHandler_UpdateNote(t *testing.T) {
	validID := uuid.New()
	now := time.Now()

	tests := []struct {
		name     string
		req      *notev1.UpdateNoteRequest
		mockRepo *MockRepository
		wantCode codes.Code
	}{
		{
			name: "valid update",
			req: &notev1.UpdateNoteRequest{
				Id:      validID.String(),
				Title:   "Updated Title",
				Content: "Updated content",
			},
			mockRepo: &MockRepository{
				GetFunc: func(ctx context.Context, id uuid.UUID) (*notedom.Note, error) {
					return &notedom.Note{
						ID:        validID,
						Title:     "Old Title",
						Content:   "Old content",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil
				},
				UpdateFunc: func(ctx context.Context, note *notedom.Note) error {
					return nil
				},
			},
			wantCode: codes.OK,
		},
		{
			name: "note not found",
			req: &notev1.UpdateNoteRequest{
				Id:    uuid.New().String(),
				Title: "New Title",
			},
			mockRepo: &MockRepository{
				GetFunc: func(ctx context.Context, id uuid.UUID) (*notedom.Note, error) {
					return nil, notedom.ErrNoteNotFound
				},
			},
			wantCode: codes.NotFound,
		},
		{
			name: "invalid uuid",
			req: &notev1.UpdateNoteRequest{
				Id:    "invalid-uuid",
				Title: "Title",
			},
			mockRepo: &MockRepository{},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "empty title",
			req: &notev1.UpdateNoteRequest{
				Id:    validID.String(),
				Title: "",
			},
			mockRepo: &MockRepository{
				GetFunc: func(ctx context.Context, id uuid.UUID) (*notedom.Note, error) {
					return &notedom.Note{
						ID:        validID,
						Title:     "Old Title",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil
				},
			},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noteuc.NewUsecase(tt.mockRepo)
			handler := NewHandler(uc)

			resp, err := handler.UpdateNote(context.Background(), tt.req)

			if tt.wantCode == codes.OK {
				if err != nil {
					t.Errorf("UpdateNote() error = %v, want nil", err)
				}
				if resp == nil || resp.Note == nil {
					t.Errorf("UpdateNote() response note missing")
				}
			} else {
				if err == nil {
					t.Errorf("UpdateNote() expected error, got nil")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantCode {
					t.Errorf("UpdateNote() code = %v, want %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

func TestHandler_DeleteNote(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name     string
		req      *notev1.DeleteNoteRequest
		mockRepo *MockRepository
		wantCode codes.Code
	}{
		{
			name: "valid delete",
			req:  &notev1.DeleteNoteRequest{Id: validID.String()},
			mockRepo: &MockRepository{
				DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			wantCode: codes.OK,
		},
		{
			name: "note not found",
			req:  &notev1.DeleteNoteRequest{Id: uuid.New().String()},
			mockRepo: &MockRepository{
				DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
					return notedom.ErrNoteNotFound
				},
			},
			wantCode: codes.NotFound,
		},
		{
			name:     "invalid uuid",
			req:      &notev1.DeleteNoteRequest{Id: "invalid-uuid"},
			mockRepo: &MockRepository{},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := noteuc.NewUsecase(tt.mockRepo)
			handler := NewHandler(uc)

			resp, err := handler.DeleteNote(context.Background(), tt.req)

			if tt.wantCode == codes.OK {
				if err != nil {
					t.Errorf("DeleteNote() error = %v, want nil", err)
				}
				if resp == nil {
					t.Errorf("DeleteNote() response is nil")
				}
			} else {
				if err == nil {
					t.Errorf("DeleteNote() expected error, got nil")
				}
				st, _ := status.FromError(err)
				if st.Code() != tt.wantCode {
					t.Errorf("DeleteNote() code = %v, want %v", st.Code(), tt.wantCode)
				}
			}
		})
	}
}

func TestMapErrorToStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{"not found", notedom.ErrNoteNotFound, codes.NotFound},
		{"empty title", notedom.ErrEmptyTitle, codes.InvalidArgument},
		{"title too long", notedom.ErrTitleTooLong, codes.InvalidArgument},
		{"unknown error", errors.New("unknown"), codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapErrorToStatus(tt.err)
			st, ok := status.FromError(result)
			if !ok {
				t.Errorf("mapErrorToStatus() did not return a gRPC status")
			}
			if st.Code() != tt.wantCode {
				t.Errorf("mapErrorToStatus() code = %v, want %v", st.Code(), tt.wantCode)
			}
		})
	}
}
