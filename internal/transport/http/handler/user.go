// Package handler provides HTTP handlers for the API.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

type createUserExecutor interface {
	Execute(ctx context.Context, req user.CreateUserRequest) (user.CreateUserResponse, error)
}

type getUserExecutor interface {
	Execute(ctx context.Context, req user.GetUserRequest) (user.GetUserResponse, error)
}

type listUsersExecutor interface {
	Execute(ctx context.Context, req user.ListUsersRequest) (user.ListUsersResponse, error)
}

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	createUC createUserExecutor
	getUC    getUserExecutor
	listUC   listUsersExecutor
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(
	createUC createUserExecutor,
	getUC getUserExecutor,
	listUC listUsersExecutor,
) *UserHandler {
	return &UserHandler{
		createUC: createUC,
		getUC:    getUC,
		listUC:   listUC,
	}
}

// CreateUser handles POST /api/v1/users.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req contract.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		contract.WriteProblemJSON(w, r, &app.AppError{
			Op:      "CreateUser",
			Code:    app.CodeValidationError,
			Message: "Invalid request body",
			Err:     err,
		})
		return
	}

	// Validate request
	if errs := contract.Validate(req); len(errs) > 0 {
		contract.WriteValidationError(w, r, errs)
		return
	}

	// Generate UUID v7 at transport boundary
	id, err := uuid.NewV7()
	if err != nil {
		contract.WriteProblemJSON(w, r, &app.AppError{
			Op:      "CreateUser",
			Code:    app.CodeInternalError,
			Message: "Failed to generate user ID",
			Err:     err,
		})
		return
	}

	// Extract context values for audit trail
	reqID := middleware.GetRequestID(r.Context())
	var actorID domain.ID
	if authCtx := app.GetAuthContext(r.Context()); authCtx != nil {
		actorID = domain.ID(authCtx.SubjectID)
	}

	// Map to app layer request
	appReq := user.CreateUserRequest{
		ID:        domain.ID(id.String()),
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RequestID: reqID,
		ActorID:   actorID,
	}

	// Execute use case
	resp, err := h.createUC.Execute(r.Context(), appReq)
	if err != nil {
		contract.WriteProblemJSON(w, r, err)
		return
	}

	// Map to response
	userResp := contract.ToUserResponse(resp.User)
	_ = contract.WriteJSON(w, http.StatusCreated, contract.DataResponse[contract.UserResponse]{Data: userResp})
}

// GetUser handles GET /api/v1/users/{id}.
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	// Validate UUID format and version
	parsedID, err := uuid.Parse(idParam)
	if err != nil {
		contract.WriteValidationError(w, r, []contract.ValidationError{
			{Field: "id", Message: "must be a valid UUID"},
		})
		return
	}
	if parsedID.Version() != 7 {
		contract.WriteValidationError(w, r, []contract.ValidationError{
			{Field: "id", Message: "must be UUID v7 (time-ordered)"},
		})
		return
	}

	// Execute use case
	resp, err := h.getUC.Execute(r.Context(), user.GetUserRequest{ID: domain.ID(parsedID.String())})
	if err != nil {
		contract.WriteProblemJSON(w, r, err)
		return
	}

	// Map to response
	userResp := contract.ToUserResponse(resp.User)
	_ = contract.WriteJSON(w, http.StatusOK, contract.DataResponse[contract.UserResponse]{Data: userResp})
}

// ListUsers handles GET /api/v1/users.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination params with defaults
	page := parseIntOrDefault(r.URL.Query().Get("page"), 1)
	pageSize := parseIntOrDefault(r.URL.Query().Get("pageSize"), 20)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // max limit
	}

	req := user.ListUsersRequest{
		Page:     page,
		PageSize: pageSize,
	}

	// Execute use case
	resp, err := h.listUC.Execute(r.Context(), req)
	if err != nil {
		contract.WriteProblemJSON(w, r, err)
		return
	}

	// Build response
	listResp := contract.NewListUsersResponse(resp.Users, page, pageSize, resp.TotalCount)
	_ = contract.WriteJSON(w, http.StatusOK, listResp)
}

func parseIntOrDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
