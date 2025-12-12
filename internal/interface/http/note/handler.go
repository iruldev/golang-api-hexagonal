package note

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/domain/note"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	noteUsecase "github.com/iruldev/golang-api-hexagonal/internal/usecase/note"
)

// Handler handles HTTP requests for Note operations.
type Handler struct {
	usecase *noteUsecase.Usecase
}

// NewHandler creates a new Note HTTP handler.
func NewHandler(u *noteUsecase.Usecase) *Handler {
	return &Handler{usecase: u}
}

// Routes registers the note routes on the given router.
func (h *Handler) Routes(r chi.Router) {
	r.Route("/notes", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// Create handles POST /api/v1/notes
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	n, err := h.usecase.Create(r.Context(), req.Title, req.Content)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.SuccessWithStatus(w, http.StatusCreated, ToNoteResponse(n))
}

// Get handles GET /api/v1/notes/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uuid, err := ParseUUID(id)
	if err != nil {
		response.BadRequest(w, "Invalid note ID format")
		return
	}

	n, err := h.usecase.Get(r.Context(), uuid)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, ToNoteResponse(n))
}

// List handles GET /api/v1/notes
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page := 1
	pageSize := 10

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	notes, total, err := h.usecase.List(r.Context(), page, pageSize)
	if err != nil {
		h.handleError(w, err)
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	response.Success(w, ListNotesResponse{
		Notes:      ToNoteResponses(notes),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// Update handles PUT /api/v1/notes/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uuid, err := ParseUUID(id)
	if err != nil {
		response.BadRequest(w, "Invalid note ID format")
		return
	}

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	n, err := h.usecase.Update(r.Context(), uuid, req.Title, req.Content)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, ToNoteResponse(n))
}

// Delete handles DELETE /api/v1/notes/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uuid, err := ParseUUID(id)
	if err != nil {
		response.BadRequest(w, "Invalid note ID format")
		return
	}

	if err := h.usecase.Delete(r.Context(), uuid); err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, nil)
}

// handleError maps domain errors to HTTP responses.
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, note.ErrNoteNotFound):
		response.NotFound(w, err.Error())
	case errors.Is(err, note.ErrEmptyTitle):
		response.ValidationError(w, err.Error())
	case errors.Is(err, note.ErrTitleTooLong):
		response.ValidationError(w, err.Error())
	default:
		response.InternalServerError(w, "An unexpected error occurred")
	}
}
