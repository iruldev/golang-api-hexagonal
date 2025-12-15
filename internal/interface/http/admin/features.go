// Package admin provides HTTP handlers for administrative endpoints.
package admin

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// FeaturesHandler provides HTTP handlers for feature flag management.
// Requires admin role (enforced at route level via RBAC middleware).
type FeaturesHandler struct {
	provider runtimeutil.AdminFeatureFlagProvider
	logger   *zap.Logger
}

// NewFeaturesHandler creates a new FeaturesHandler.
// The provider is required; logger is optional (can be nil).
func NewFeaturesHandler(provider runtimeutil.AdminFeatureFlagProvider, logger *zap.Logger) *FeaturesHandler {
	return &FeaturesHandler{
		provider: provider,
		logger:   logger,
	}
}

// ListFlags handles GET /admin/features
// Returns a list of all configured feature flags.
//
// Response:
//
//	{
//	  "success": true,
//	  "data": [
//	    {"name": "new_dashboard", "enabled": true, "description": "New dashboard UI", "updated_at": "2025-12-14T22:00:00Z"}
//	  ]
//	}
func (h *FeaturesHandler) ListFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.provider.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to list feature flags")
		return
	}

	response.Success(w, flags)
}

// GetFlag handles GET /admin/features/{flag}
// Returns details of a specific feature flag.
//
// Response (200):
//
//	{
//	  "success": true,
//	  "data": {"name": "new_dashboard", "enabled": true, "description": "...", "updated_at": "..."}
//	}
//
// Response (404):
//
//	{
//	  "success": false,
//	  "error": {"code": "ERR_NOT_FOUND", "message": "Feature flag not found"}
//	}
func (h *FeaturesHandler) GetFlag(w http.ResponseWriter, r *http.Request) {
	flagName := chi.URLParam(r, "flag")

	state, err := h.provider.Get(r.Context(), flagName)
	if err != nil {
		if errors.Is(err, runtimeutil.ErrFlagNotFound) {
			response.Error(w, http.StatusNotFound, "ERR_NOT_FOUND", "Feature flag not found")
			return
		}
		if errors.Is(err, runtimeutil.ErrInvalidFlagName) {
			response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid flag name")
			return
		}
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to get feature flag")
		return
	}

	response.Success(w, state)
}

// EnableFlag handles POST /admin/features/{flag}/enable
// Enables a feature flag.
//
// Response (201):
//
//	{
//	  "success": true,
//	  "data": {"flag": "new_dashboard", "enabled": true}
//	}
func (h *FeaturesHandler) EnableFlag(w http.ResponseWriter, r *http.Request) {
	h.setFlagState(w, r, true)
}

// DisableFlag handles POST /admin/features/{flag}/disable
// Disables a feature flag.
//
// Response (201):
//
//	{
//	  "success": true,
//	  "data": {"flag": "new_dashboard", "enabled": false}
//	}
func (h *FeaturesHandler) DisableFlag(w http.ResponseWriter, r *http.Request) {
	h.setFlagState(w, r, false)
}

// setFlagState is a helper that enables or disables a flag.
func (h *FeaturesHandler) setFlagState(w http.ResponseWriter, r *http.Request, enabled bool) {
	flagName := chi.URLParam(r, "flag")

	// Validate flag name
	if flagName == "" {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Flag name is required")
		return
	}

	// Get actor from claims for audit logging
	claims, _ := ctxutil.ClaimsFromContext(r.Context())
	actorID := claims.UserID
	if actorID == "" {
		actorID = "unknown"
	}

	// Update flag state
	if err := h.provider.SetEnabled(r.Context(), flagName, enabled); err != nil {
		if errors.Is(err, runtimeutil.ErrInvalidFlagName) {
			response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid flag name")
			return
		}
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to update feature flag")
		return
	}

	// Audit logging
	action := observability.ActionUpdate
	actionDetails := "disable"
	if enabled {
		actionDetails = "enable"
	}

	auditEvent := observability.NewAuditEvent(
		r.Context(),
		action,
		"feature_flag:"+flagName,
		actorID,
		map[string]any{
			"action_type": actionDetails,
			"enabled":     enabled,
		},
	)
	observability.LogAudit(r.Context(), h.logger, auditEvent)

	// Return response with 201 Created
	response.SuccessWithStatus(w, http.StatusCreated, map[string]interface{}{
		"flag":    flagName,
		"enabled": enabled,
	})
}
