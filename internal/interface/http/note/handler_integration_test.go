//go:build integration
// +build integration

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
)

// TestNoteHandler_Integration tests the Note HTTP handlers end-to-end.
// Run with: go test -tags=integration ./...
// Or skip with: go test -short ./...
func TestNoteHandler_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup: Create mock repository (in real scenario, use test database)
	repo := newTestRepository()
	uc := noteUsecase.NewUsecase(repo)
	handler := NewHandler(uc)

	// Create router with note routes
	router := chi.NewRouter()
	router.Route("/api/v1", func(r chi.Router) {
		handler.Routes(r)
	})

	// Create test server
	srv := httptest.NewServer(router)
	defer srv.Close()

	client := srv.Client()
	baseURL := srv.URL + "/api/v1"

	var createdNoteID string

	// Test Create
	t.Run("POST /notes - create note", func(t *testing.T) {
		body := `{"title":"Integration Test Note","content":"Test content for integration"}`
		resp, err := client.Post(baseURL+"/notes", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected status 201, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["success"] != true {
			t.Error("expected success=true")
		}

		data := result["data"].(map[string]interface{})
		createdNoteID = data["id"].(string)
		if createdNoteID == "" {
			t.Error("expected note ID")
		}
	})

	// Test Get
	t.Run("GET /notes/:id - get note", func(t *testing.T) {
		if createdNoteID == "" {
			t.Skip("no note created")
		}

		resp, err := client.Get(baseURL + "/notes/" + createdNoteID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["success"] != true {
			t.Error("expected success=true")
		}
	})

	// Test List
	t.Run("GET /notes - list notes", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/notes?page=1&page_size=10")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["success"] != true {
			t.Error("expected success=true")
		}
	})

	// Test Update
	t.Run("PUT /notes/:id - update note", func(t *testing.T) {
		if createdNoteID == "" {
			t.Skip("no note created")
		}

		body := `{"title":"Updated Title","content":"Updated content"}`
		req, _ := http.NewRequest(http.MethodPut, baseURL+"/notes/"+createdNoteID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	// Test Delete
	t.Run("DELETE /notes/:id - delete note", func(t *testing.T) {
		if createdNoteID == "" {
			t.Skip("no note created")
		}

		req, _ := http.NewRequest(http.MethodDelete, baseURL+"/notes/"+createdNoteID, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	// Cleanup: Verify note is deleted
	t.Run("GET /notes/:id - verify deleted", func(t *testing.T) {
		if createdNoteID == "" {
			t.Skip("no note created")
		}

		resp, err := client.Get(baseURL + "/notes/" + createdNoteID)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})
}

// testRepository is an in-memory repository for integration tests.
// In production, this would connect to a test database.
type testRepository struct {
	notes map[uuid.UUID]*note.Note
}

func newTestRepository() *testRepository {
	return &testRepository{
		notes: make(map[uuid.UUID]*note.Note),
	}
}

func (r *testRepository) Create(ctx context.Context, n *note.Note) error {
	r.notes[n.ID] = n
	return nil
}

func (r *testRepository) Get(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	n, ok := r.notes[id]
	if !ok {
		return nil, note.ErrNoteNotFound
	}
	return n, nil
}

func (r *testRepository) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	notes := make([]*note.Note, 0, len(r.notes))
	for _, n := range r.notes {
		notes = append(notes, n)
	}
	total := int64(len(notes))

	// Apply pagination
	if offset >= len(notes) {
		return []*note.Note{}, total, nil
	}
	end := offset + limit
	if end > len(notes) {
		end = len(notes)
	}

	return notes[offset:end], total, nil
}

func (r *testRepository) Update(ctx context.Context, n *note.Note) error {
	if _, ok := r.notes[n.ID]; !ok {
		return note.ErrNoteNotFound
	}
	r.notes[n.ID] = n
	return nil
}

func (r *testRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := r.notes[id]; !ok {
		return note.ErrNoteNotFound
	}
	delete(r.notes, id)
	return nil
}
