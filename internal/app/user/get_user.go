// Package user provides use cases for user-related operations.
// This package follows hexagonal architecture principles - it depends only on the domain layer
// and defines ports (interfaces) that infrastructure adapters must implement.
package user

import (
	"context"
	"errors"
	"strings"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/logger"
)

// GetUserRequest represents the input data for getting a user by ID.
type GetUserRequest struct {
	ID domain.ID
}

// GetUserResponse represents the result of getting a user.
type GetUserResponse struct {
	User domain.User
}

// OpGetUser is the operation name for GetUser use case.
// It is used for error wrapping and logging in the GetUser use case.
const OpGetUser = "GetUser"

// GetUserUseCase handles the business logic for retrieving a user by ID.
// It demonstrates the app-layer authorization pattern per architecture.md:
// - Authorization checks happen at START of use case, before any DB calls.
// - Admins can get any user, regular users can only get their own profile.
// Story 2.8: Includes audit logging for authorization decisions.

type GetUserUseCase struct {
	userRepo domain.UserRepository
	db       domain.Querier
	log      *logger.Logger
}

// NewGetUserUseCase creates a new instance of GetUserUseCase.
func NewGetUserUseCase(userRepo domain.UserRepository, db domain.Querier, log *logger.Logger) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
		db:       db,
		log:      log.With("usecase", "GetUser"),
	}
}

// Execute retrieves a user by ID.
// Authorization: Admins can get any user, regular users can only get themselves.
// Returns AppError with Code=FORBIDDEN if:
// - No auth context is present (fail-closed for protected routes)
// - User tries to access another user's data
// Returns AppError with Code=USER_NOT_FOUND if user doesn't exist.
func (uc *GetUserUseCase) Execute(ctx context.Context, req GetUserRequest) (GetUserResponse, error) {
	// Authorization check at START of use case (per architecture.md)
	// This happens BEFORE any database calls for fail-fast behavior.
	authCtx := app.GetAuthContext(ctx)

	// Fail-closed: If no auth context is present, deny access
	// This enforces AC #1 - protected routes require authentication
	// Also fail-closed for missing or unknown roles.
	if authCtx == nil || strings.TrimSpace(authCtx.Role) == "" || strings.TrimSpace(authCtx.SubjectID) == "" {
		// Story 2.8: Audit log authorization denial
		logger.FromContext(ctx, uc.log).WarnContext(ctx, "authorization denied: no auth context or invalid credentials",
			"resourceId", req.ID,
		)
		return GetUserResponse{}, &app.AppError{
			Op:      OpGetUser,
			Code:    app.CodeForbidden,
			Message: "Access denied",
			Err:     app.ErrNoAuthContext,
		}
	}

	// Only known roles are allowed; unknown roles are denied even if subject matches.
	if !authCtx.IsAdmin() && !authCtx.IsUser() {
		// Story 2.8: Audit log unknown role denial
		logger.FromContext(ctx, uc.log).WarnContext(ctx, "authorization denied: unknown role",
			"actorId", authCtx.SubjectID,
			"role", authCtx.Role,
			"resourceId", req.ID,
		)
		return GetUserResponse{}, &app.AppError{
			Op:      OpGetUser,
			Code:    app.CodeForbidden,
			Message: "Access denied",
		}
	}

	// Enforce authorization rules:
	// - Admins can access any user
	// - Regular users can only access their own profile
	if authCtx.IsUser() && authCtx.SubjectID != string(req.ID) {
		// Story 2.8: Audit log IDOR attempt
		logger.FromContext(ctx, uc.log).WarnContext(ctx, "authorization denied: IDOR attempt",
			"actorId", authCtx.SubjectID,
			"resourceId", req.ID,
		)
		return GetUserResponse{}, &app.AppError{
			Op:      OpGetUser,
			Code:    app.CodeForbidden,
			Message: "Access denied",
		}
	}

	// Story 2.8: Audit log authorization granted
	logger.FromContext(ctx, uc.log).DebugContext(ctx, "authorization granted",
		"actorId", authCtx.SubjectID,
		"role", authCtx.Role,
		"resourceId", req.ID,
	)

	user, err := uc.userRepo.GetByID(ctx, uc.db, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return GetUserResponse{}, &app.AppError{
				Op:      OpGetUser,
				Code:    app.CodeUserNotFound,
				Message: "User not found",
				Err:     err,
			}
		}
		return GetUserResponse{}, &app.AppError{
			Op:      OpGetUser,
			Code:    app.CodeInternalError,
			Message: "Failed to get user",
			Err:     err,
		}
	}
	if user == nil {
		return GetUserResponse{}, &app.AppError{
			Op:      "GetUser",
			Code:    app.CodeInternalError,
			Message: "Failed to get user",
			Err:     errors.New("user repository returned nil user without error"),
		}
	}
	return GetUserResponse{User: *user}, nil
}
