// Package handler provides HTTP handlers for the API.
package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
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
	createUC     createUserExecutor
	getUC        getUserExecutor
	listUC       listUsersExecutor
	resourcePath string
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(
	createUC createUserExecutor,
	getUC getUserExecutor,
	listUC listUsersExecutor,
	resourcePath string,
) *UserHandler {
	return &UserHandler{
		createUC:     createUC,
		getUC:        getUC,
		listUC:       listUC,
		resourcePath: resourcePath,
	}
}

// CreateUser handles POST /api/v1/users.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req contract.CreateUserRequest
	// Decode and validate request
	if errs := contract.ValidateRequestBody(r, &req); len(errs) > 0 {
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
	reqID := ctxutil.GetRequestID(r.Context())
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

	// Set Location header for 201 Created (before writing response body)
	location := fmt.Sprintf("%s/%s", h.resourcePath, resp.User.ID)
	w.Header().Set("Location", location)

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
	// Parse pagination params
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	// Defaults
	page := 1
	pageSize := 20

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			contract.WriteValidationError(w, r, []contract.ValidationError{
				{Field: "page", Message: "must be a positive integer", Code: contract.CodeValOutOfRange},
			})
			return
		}
		page = p
	}

	if pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil || ps < 1 {
			contract.WriteValidationError(w, r, []contract.ValidationError{
				{Field: "pageSize", Message: "must be a positive integer", Code: contract.CodeValOutOfRange},
			})
			return
		}
		if ps > 100 {
			// Cap at 100 but don't error, as per common practice?
			// Spec says: "Maximum: 100 (values above 100 are capped)"
			// So we just cap it.
			ps = 100
		}
		pageSize = ps
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
