package graphql_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/iruldev/golang-api-hexagonal/internal/domain/note"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/graphql"
	noteuc "github.com/iruldev/golang-api-hexagonal/internal/usecase/note"
)

// MockRepository is a mock implementation of note.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, n *note.Note) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockRepository) Get(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, n *note.Note) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*note.Note), args.Get(1).(int64), args.Error(2)
}

func TestGraphQL_Integration(t *testing.T) {
	mockRepo := new(MockRepository)
	usecase := noteuc.NewUsecase(mockRepo)

	resolver := &graphql.Resolver{
		NoteUsecase: usecase,
	}

	srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{
		Resolvers: resolver,
	}))

	t.Run("CreateNote", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*note.Note")).Return(nil)

		q := `mutation { createNote(input: {title: "New Note", content: "New Content"}) { title content } }`
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", createBody(q))
		r.Header.Set("Content-Type", "application/json")

		srv.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "New Note")
	})

	t.Run("UpdateNote", func(t *testing.T) {
		noteID := uuid.New()

		// Create a note and force its ID to match noteID
		existingNote := note.NewNote("Old", "Old")
		existingNote.ID = noteID

		// Mock Get to return this specific note instance
		mockRepo.On("Get", mock.Anything, noteID).Return(existingNote, nil)

		// Mock Update to expect the SAME instance (or matched by ID)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(n *note.Note) bool {
			return n.ID == noteID && n.Title == "Updated"
		})).Return(nil)

		q := `mutation { updateNote(id: "` + noteID.String() + `", input: {title: "Updated", content: "Updated Content"}) { title } }`
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", createBody(q))
		r.Header.Set("Content-Type", "application/json")

		srv.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Updated")
	})

	t.Run("DeleteNote", func(t *testing.T) {
		noteID := uuid.New()
		// Get is called first usually? delete might check existence?
		// NoteUsecase.Delete doesn't call Get? checking usecase.go...
		// "func (u *Usecase) Delete(ctx context.Context, id uuid.UUID) error { return u.repo.Delete(ctx, id) }"
		// So it just calls Repo.Delete.
		// But wait, the test setup I wrote earlier had: mockRepo.On("Get", ...).Return(...)
		// I should check if Delete calls Get.
		// If it doesn't, attempting to Get will cause "Unexpected call".
		// But in the previous run, Delete PASSED!
		// "--- PASS: TestGraphQL_Integration/DeleteNote (0.00s)"
		// So Usecase.Delete probably DOES NOT call Get.
		// And my previous test setup included: mockRepo.On("Get", ...).
		// Why didn't it fail with "Missing call"?
		// Ah, mockery strictness? usually "On" setups that aren't called might be ignored unless Strict expectation?
		// Or maybe Usecase.Delete DOES call Get?
		// Checking usecase.go again... Line 84: return u.repo.Delete(ctx, id)
		// It does NOT call Get.
		// So "Get" expectation is superfluous.
		// I will remove it to be clean.

		mockRepo.On("Delete", mock.Anything, noteID).Return(nil)

		q := `mutation { deleteNote(id: "` + noteID.String() + `") }`
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", createBody(q))
		r.Header.Set("Content-Type", "application/json")

		srv.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "true")
	})

	t.Run("GetNote", func(t *testing.T) {
		noteID := uuid.New()
		mockRepo.On("Get", mock.Anything, noteID).Return(note.NewNote("My Note", "Content"), nil)

		q := `{ note(id: "` + noteID.String() + `") { title } }`
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", createBody(q))
		r.Header.Set("Content-Type", "application/json")

		srv.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "My Note")
	})

	t.Run("GetNote_NotFound", func(t *testing.T) {
		noteID := uuid.New()
		mockRepo.On("Get", mock.Anything, noteID).Return(nil, note.ErrNoteNotFound)

		q := `{ note(id: "` + noteID.String() + `") { title } }`
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", createBody(q))
		r.Header.Set("Content-Type", "application/json")

		srv.ServeHTTP(w, r)

		// GraphQL returns 200 even on errors, but payload has errors
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "NOT_FOUND")
	})

	t.Run("ListNotes", func(t *testing.T) {
		notes := []*note.Note{
			note.NewNote("List Note", "List Content"),
		}
		mockRepo.On("List", mock.Anything, 100, 0).Return(notes, int64(1), nil)

		q := `{ notes { title } }`
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/query", createBody(q))
		r.Header.Set("Content-Type", "application/json")

		srv.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "List Note")
	})
}

func createBody(query string) *strings.Reader {
	body, _ := json.Marshal(map[string]string{
		"query": query,
	})
	return strings.NewReader(string(body))
}
