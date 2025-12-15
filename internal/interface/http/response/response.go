// Package response provides HTTP response helpers for consistent API responses.
package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// SuccessResponse represents a successful API response.
// Used for all 2xx responses.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an error API response.
// Used for all 4xx and 5xx responses.
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail contains error information.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON writes a JSON response with the given status code.
// Sets Content-Type to application/json.
// Note: If encoding fails, the error is logged but status cannot be changed
// since WriteHeader has already been called.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log encoding errors - can't change status after WriteHeader
		// Recovery middleware will catch panics if they occur
		log.Printf("response: failed to encode JSON: %v", err)
	}
}

// Success writes a success response with HTTP 200 OK.
// Response format: {"success": true, "data": {...}}
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithStatus writes a success response with a custom HTTP status.
// Useful for 201 Created, 202 Accepted, etc.
// Deprecated: Use SuccessEnvelopeWithStatus instead.
func SuccessWithStatus(w http.ResponseWriter, status int, data interface{}) {
	JSON(w, status, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// Error writes an error response with the given HTTP status.
// Response format: {"success": false, "error": {"code": "...", "message": "..."}}
// Deprecated: Use ErrorEnvelope instead.
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// BadRequest writes a 400 Bad Request error response.
// Deprecated: Use BadRequestCtx instead.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, ErrBadRequest, message)
}

// Unauthorized writes a 401 Unauthorized error response.
// Deprecated: Use UnauthorizedCtx instead.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, ErrUnauthorized, message)
}

// Forbidden writes a 403 Forbidden error response.
// Deprecated: Use ForbiddenCtx instead.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, ErrForbidden, message)
}

// NotFound writes a 404 Not Found error response.
// Deprecated: Use NotFoundCtx instead.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, ErrNotFound, message)
}

// Conflict writes a 409 Conflict error response.
// Deprecated: Use ConflictCtx instead.
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, ErrConflict, message)
}

// ValidationError writes a 422 Unprocessable Entity error response.
// Deprecated: Use ValidationErrorCtx instead.
func ValidationError(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnprocessableEntity, ErrValidation, message)
}

// InternalServerError writes a 500 Internal Server Error response.
// Deprecated: Use InternalServerErrorCtx instead.
func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "ERR_INTERNAL_SERVER", message)
}
