// Package admin provides HTTP handlers for administrative endpoints.
package admin

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// SetRolesRequest is the request body for setting user roles.
type SetRolesRequest struct {
	Roles []string `json:"roles"`
}

// ModifyRoleRequest is the request body for adding or removing a single role.
type ModifyRoleRequest struct {
	Role string `json:"role"`
}

// UserRolesResponse is the response body for user role operations.
type UserRolesResponse struct {
	UserID    string   `json:"user_id"`
	Roles     []string `json:"roles"`
	UpdatedAt string   `json:"updated_at"`
}

// RolesHandler provides HTTP handlers for user role management.
// Requires admin role (enforced at route level via RBAC middleware).
type RolesHandler struct {
	provider runtimeutil.UserRoleProvider
	logger   *zap.Logger
}

// NewRolesHandler creates a new RolesHandler.
// The provider is required; logger is optional (can be nil).
func NewRolesHandler(provider runtimeutil.UserRoleProvider, logger *zap.Logger) *RolesHandler {
	return &RolesHandler{
		provider: provider,
		logger:   logger,
	}
}

// toResponse converts UserRoles to UserRolesResponse.
func toResponse(roles *runtimeutil.UserRoles) UserRolesResponse {
	updatedAt := ""
	if !roles.UpdatedAt.IsZero() {
		updatedAt = roles.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return UserRolesResponse{
		UserID:    roles.UserID,
		Roles:     roles.Roles,
		UpdatedAt: updatedAt,
	}
}

// validateUserID validates that the user ID is a valid UUID.
func validateUserID(userID string) error {
	_, err := uuid.Parse(userID)
	return err
}

// GetUserRoles handles GET /admin/users/{id}/roles
// Returns the roles assigned to a user.
//
// Response (200):
//
//	{
//	  "success": true,
//	  "data": {
//	    "user_id": "550e8400-e29b-41d4-a716-446655440000",
//	    "roles": ["admin", "user"],
//	    "updated_at": "2025-12-14T23:00:00Z"
//	  }
//	}
func (h *RolesHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Validate UUID format
	if err := validateUserID(userID); err != nil {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid user ID format")
		return
	}

	roles, err := h.provider.GetUserRoles(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get user roles for %s: %v", userID, err)
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to get user roles")
		return
	}

	response.Success(w, toResponse(roles))
}

// SetUserRoles handles POST /admin/users/{id}/roles
// Replaces all roles for a user.
//
// Request:
//
//	{"roles": ["admin", "user"]}
//
// Response (200):
//
//	{
//	  "success": true,
//	  "data": {
//	    "user_id": "550e8400-e29b-41d4-a716-446655440000",
//	    "roles": ["admin", "user"],
//	    "updated_at": "2025-12-14T23:00:00Z"
//	  }
//	}
func (h *RolesHandler) SetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Validate UUID format
	if err := validateUserID(userID); err != nil {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid user ID format")
		return
	}

	// Parse request body
	var req SetRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid request body")
		return
	}

	// Get old roles for audit logging
	oldRoles, _ := h.provider.GetUserRoles(r.Context(), userID)
	oldRolesList := []string{}
	if oldRoles != nil {
		oldRolesList = oldRoles.Roles
	}

	// Set roles
	newRoles, err := h.provider.SetUserRoles(r.Context(), userID, req.Roles)
	if err != nil {
		if err == runtimeutil.ErrInvalidRole {
			response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid role name")
			return
		}
		log.Printf("Failed to set user roles for %s: %v", userID, err)
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to set user roles")
		return
	}

	// Audit logging
	h.logRoleChange(r, userID, "set_roles", oldRolesList, newRoles.Roles)

	response.Success(w, toResponse(newRoles))
}

// AddUserRole handles POST /admin/users/{id}/roles/add
// Adds a single role to a user.
//
// Request:
//
//	{"role": "admin"}
//
// Response (200):
//
//	{
//	  "success": true,
//	  "data": {
//	    "user_id": "550e8400-e29b-41d4-a716-446655440000",
//	    "roles": ["user", "admin"],
//	    "updated_at": "2025-12-14T23:00:00Z"
//	  }
//	}
func (h *RolesHandler) AddUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Validate UUID format
	if err := validateUserID(userID); err != nil {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid user ID format")
		return
	}

	// Parse request body
	var req ModifyRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid request body")
		return
	}

	if req.Role == "" {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Role is required")
		return
	}

	// Get old roles for audit logging
	oldRoles, _ := h.provider.GetUserRoles(r.Context(), userID)
	oldRolesList := []string{}
	if oldRoles != nil {
		oldRolesList = oldRoles.Roles
	}

	// Add role
	newRoles, err := h.provider.AddUserRole(r.Context(), userID, req.Role)
	if err != nil {
		if err == runtimeutil.ErrInvalidRole {
			response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid role name")
			return
		}
		log.Printf("Failed to add role for user %s: %v", userID, err)
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to add user role")
		return
	}

	// Audit logging
	h.logRoleChange(r, userID, "add_role", oldRolesList, newRoles.Roles)

	response.Success(w, toResponse(newRoles))
}

// RemoveUserRole handles POST /admin/users/{id}/roles/remove
// Removes a single role from a user.
//
// Request:
//
//	{"role": "admin"}
//
// Response (200):
//
//	{
//	  "success": true,
//	  "data": {
//	    "user_id": "550e8400-e29b-41d4-a716-446655440000",
//	    "roles": ["user"],
//	    "updated_at": "2025-12-14T23:00:00Z"
//	  }
//	}
func (h *RolesHandler) RemoveUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Validate UUID format
	if err := validateUserID(userID); err != nil {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid user ID format")
		return
	}

	// Parse request body
	var req ModifyRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid request body")
		return
	}

	if req.Role == "" {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Role is required")
		return
	}

	// Get old roles for audit logging
	oldRoles, _ := h.provider.GetUserRoles(r.Context(), userID)
	oldRolesList := []string{}
	if oldRoles != nil {
		oldRolesList = oldRoles.Roles
	}

	// Remove role
	newRoles, err := h.provider.RemoveUserRole(r.Context(), userID, req.Role)
	if err != nil {
		if err == runtimeutil.ErrInvalidRole {
			response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid role name")
			return
		}
		log.Printf("Failed to remove role for user %s: %v", userID, err)
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to remove user role")
		return
	}

	// Audit logging
	h.logRoleChange(r, userID, "remove_role", oldRolesList, newRoles.Roles)

	response.Success(w, toResponse(newRoles))
}

// logRoleChange logs a role change event with audit logging.
func (h *RolesHandler) logRoleChange(r *http.Request, userID, actionType string, oldRoles, newRoles []string) {
	// Get actor from claims
	claims, err := ctxutil.ClaimsFromContext(r.Context())
	actorID := claims.UserID
	if actorID == "" {
		actorID = "unknown"
	}
	if err != nil {
		log.Printf("Warning: could not get claims from context for audit log: %v", err)
	}

	// Determine action type
	action := observability.ActionUpdate
	if actionType == "add_role" {
		action = observability.ActionCreate
	} else if actionType == "remove_role" {
		action = observability.ActionDelete
	}

	auditEvent := observability.NewAuditEvent(
		r.Context(),
		action,
		"user_role:"+userID,
		actorID,
		map[string]any{
			"action_type": actionType,
			"old_roles":   oldRoles,
			"new_roles":   newRoles,
		},
	)
	observability.LogAudit(r.Context(), h.logger, auditEvent)
}
