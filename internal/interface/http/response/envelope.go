// Package response provides HTTP response helpers for consistent API responses.
package response

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
)

// UnknownTraceID is used when trace_id cannot be extracted from context.
const UnknownTraceID = "unknown"

// Envelope represents a standard API response structure.
// All API responses follow this format for consistency.
type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
	Meta  *Meta      `json:"meta,omitempty"`
}

// Meta contains response metadata including trace information.
// trace_id is MANDATORY and must be present in all responses.
type Meta struct {
	TraceID  string `json:"trace_id"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Total    int    `json:"total,omitempty"`
}

// ErrorBody contains error information for error responses.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// Hint provides additional guidance.
	// WARNING: Do not include sensitive info or internal error details here.
	Hint string `json:"hint,omitempty"`
}

// Pagination contains pagination parameters.
// Used to construct paginated responses.
type Pagination struct {
	Page     int
	PageSize int
	Total    int
}

// getTraceID extracts trace_id from context or returns fallback.
func getTraceID(ctx context.Context) string {
	if traceID := ctxutil.RequestIDFromContext(ctx); traceID != "" {
		return traceID
	}
	return UnknownTraceID
}

// newMeta creates a Meta struct with trace_id from context.
func newMeta(ctx context.Context) *Meta {
	return &Meta{
		TraceID: getTraceID(ctx),
	}
}

// newMetaWithPagination creates a Meta struct with pagination info.
func newMetaWithPagination(ctx context.Context, p Pagination) *Meta {
	return &Meta{
		TraceID:  getTraceID(ctx),
		Page:     p.Page,
		PageSize: p.PageSize,
		Total:    p.Total,
	}
}

// WriteJSON writes a JSON response with the given status code.
// Sets Content-Type to application/json.
// Note: If encoding fails, the error is logged but status cannot be changed
// since WriteHeader has already been called.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log encoding errors - can't change status after WriteHeader
		slog.Error("response: failed to encode JSON", "error", err)
	}
}

// SuccessEnvelope writes a success response with HTTP 200 OK.
// Response format: {"data": {...}, "meta": {"trace_id": "..."}}
func SuccessEnvelope(w http.ResponseWriter, ctx context.Context, data any) {
	WriteJSON(w, http.StatusOK, Envelope{
		Data: data,
		Meta: newMeta(ctx),
	})
}

// SuccessEnvelopeWithStatus writes a success response with a custom HTTP status.
// Useful for 201 Created, 202 Accepted, etc.
func SuccessEnvelopeWithStatus(w http.ResponseWriter, status int, ctx context.Context, data any) {
	WriteJSON(w, status, Envelope{
		Data: data,
		Meta: newMeta(ctx),
	})
}

// SuccessEnvelopeWithPagination writes a paginated success response.
// Response format: {"data": [...], "meta": {"trace_id": "...", "page": 1, "page_size": 10, "total": 100}}
func SuccessEnvelopeWithPagination(w http.ResponseWriter, ctx context.Context, data any, p Pagination) {
	WriteJSON(w, http.StatusOK, Envelope{
		Data: data,
		Meta: newMetaWithPagination(ctx, p),
	})
}

// ErrorEnvelope writes an error response with the given HTTP status.
// Response format: {"error": {"code": "...", "message": "..."}, "meta": {"trace_id": "..."}}
func ErrorEnvelope(w http.ResponseWriter, ctx context.Context, status int, code, message string) {
	WriteJSON(w, status, Envelope{
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
		Meta: newMeta(ctx),
	})
}

// ErrorEnvelopeWithHint writes an error response with an optional hint.
// Hints provide additional guidance on how to resolve the error.
func ErrorEnvelopeWithHint(w http.ResponseWriter, ctx context.Context, status int, code, message, hint string) {
	WriteJSON(w, status, Envelope{
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Hint:    hint,
		},
		Meta: newMeta(ctx),
	})
}

// BadRequestCtx writes a 400 Bad Request error response with context.
func BadRequestCtx(w http.ResponseWriter, ctx context.Context, message string) {
	ErrorEnvelope(w, ctx, http.StatusBadRequest, CodeBadRequest, message)
}

// UnauthorizedCtx writes a 401 Unauthorized error response with context.
func UnauthorizedCtx(w http.ResponseWriter, ctx context.Context, message string) {
	ErrorEnvelope(w, ctx, http.StatusUnauthorized, CodeUnauthorized, message)
}

// ForbiddenCtx writes a 403 Forbidden error response with context.
func ForbiddenCtx(w http.ResponseWriter, ctx context.Context, message string) {
	ErrorEnvelope(w, ctx, http.StatusForbidden, CodeForbidden, message)
}

// NotFoundCtx writes a 404 Not Found error response with context.
func NotFoundCtx(w http.ResponseWriter, ctx context.Context, message string) {
	ErrorEnvelope(w, ctx, http.StatusNotFound, CodeNotFound, message)
}

// ConflictCtx writes a 409 Conflict error response with context.
func ConflictCtx(w http.ResponseWriter, ctx context.Context, message string) {
	ErrorEnvelope(w, ctx, http.StatusConflict, CodeConflict, message)
}

// ValidationErrorCtx writes a 422 Unprocessable Entity error response with context.
func ValidationErrorCtx(w http.ResponseWriter, ctx context.Context, message string) {
	ErrorEnvelope(w, ctx, http.StatusUnprocessableEntity, CodeValidation, message)
}

// InternalServerErrorCtx writes a 500 Internal Server Error response with context.
func InternalServerErrorCtx(w http.ResponseWriter, ctx context.Context, message string) {
	ErrorEnvelope(w, ctx, http.StatusInternalServerError, CodeInternalServer, message)
}

// ServiceUnavailableCtx writes a 503 Service Unavailable error response with context.
func ServiceUnavailableCtx(w http.ResponseWriter, ctx context.Context, message string) {
	ErrorEnvelope(w, ctx, http.StatusServiceUnavailable, CodeServiceUnavailable, message)
}

// TimeoutCtx writes a 504 Gateway Timeout error response with context.
func TimeoutCtx(w http.ResponseWriter, ctx context.Context, message string) {
	ErrorEnvelope(w, ctx, http.StatusGatewayTimeout, CodeTimeout, message)
}
