package note

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/iruldev/golang-api-hexagonal/internal/domain/note"
	noteUsecase "github.com/iruldev/golang-api-hexagonal/internal/usecase/note"
	"go.uber.org/zap"
)

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		statusCode int
	}{
		{
			name:       "success",
			body:       `{"title":"Test Title","content":"Test Content"}`,
			statusCode: http.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `invalid`,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockRepo{}
			uc := noteUsecase.NewUsecase(repo, zap.NewNop())
			handler := NewHandler(uc)

			req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			handler.Create(w, req)

			// Assert
			if w.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		repoErr    error
		statusCode int
	}{
		{
			name:       "success",
			id:         uuid.New().String(),
			statusCode: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			id:         "invalid",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "not found",
			id:         uuid.New().String(),
			repoErr:    note.ErrNoteNotFound,
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockRepo{getErr: tt.repoErr}
			uc := noteUsecase.NewUsecase(repo, zap.NewNop())
			handler := NewHandler(uc)

			req := httptest.NewRequest(http.MethodGet, "/notes/"+tt.id, nil)
			w := httptest.NewRecorder()

			// Add chi URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Act
			handler.Get(w, req)

			// Assert
			if w.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestHandler_List(t *testing.T) {
	// Arrange
	repo := &mockRepo{}
	uc := noteUsecase.NewUsecase(repo, zap.NewNop())
	handler := NewHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/notes?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	// Act
	handler.List(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["success"] != true {
		t.Error("expected success=true")
	}
}

func TestHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		repoErr    error
		statusCode int
	}{
		{
			name:       "success",
			id:         uuid.New().String(),
			body:       `{"title":"Updated","content":"Updated Content"}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			id:         "invalid",
			body:       `{"title":"Test","content":"Content"}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			id:         uuid.New().String(),
			body:       `invalid`,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockRepo{getErr: tt.repoErr}
			uc := noteUsecase.NewUsecase(repo, zap.NewNop())
			handler := NewHandler(uc)

			req := httptest.NewRequest(http.MethodPut, "/notes/"+tt.id, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Act
			handler.Update(w, req)

			// Assert
			if w.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		repoErr    error
		statusCode int
	}{
		{
			name:       "success",
			id:         uuid.New().String(),
			statusCode: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			id:         "invalid",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "not found",
			id:         uuid.New().String(),
			repoErr:    note.ErrNoteNotFound,
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mockRepo{deleteErr: tt.repoErr}
			uc := noteUsecase.NewUsecase(repo, zap.NewNop())
			handler := NewHandler(uc)

			// Act
			req := httptest.NewRequest(http.MethodDelete, "/notes/"+tt.id, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.Delete(w, req)

			// Assert
			if w.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

// mockRepo is a simple mock for testing handlers.
type mockRepo struct {
	getErr    error
	deleteErr error
}

func (m *mockRepo) Create(_ context.Context, n *note.Note) error {
	return nil
}

func (m *mockRepo) Get(_ context.Context, id uuid.UUID) (*note.Note, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return note.NewNote("Test", "Content"), nil
}

func (m *mockRepo) List(_ context.Context, limit, offset int) ([]*note.Note, int64, error) {
	return []*note.Note{note.NewNote("Test", "Content")}, 1, nil
}

func (m *mockRepo) Update(_ context.Context, n *note.Note) error {
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	return m.deleteErr
}

type Context = context.Context
