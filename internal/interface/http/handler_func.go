// Package http provides HTTP interface layer components.
package http

import (
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// HandlerFuncE is a handler function that can return an error.
// This enables cleaner handler code without explicit error handling.
//
// Usage:
//
//	func GetUser(w http.ResponseWriter, r *http.Request) error {
//	    user, err := service.GetUser(r.Context(), id)
//	    if err != nil {
//	        return err  // Automatically mapped to Envelope response
//	    }
//	    response.SuccessEnvelope(w, r.Context(), user)
//	    return nil
//	}
//
//	router.Get("/users/{id}", http.WrapHandler(GetUser))
type HandlerFuncE func(w http.ResponseWriter, r *http.Request) error

// WrapHandler converts a HandlerFuncE to http.HandlerFunc.
// Errors returned by the handler are automatically mapped to Envelope responses
// using response.HandleErrorCtx which handles:
//   - DomainError with code, message, and hint
//   - Legacy sentinel errors for backward compatibility
//   - Unknown errors as 500 Internal Server Error
//
// The wrapper ensures:
//   - Consistent error response format (Envelope)
//   - meta.trace_id is included in all error responses
//   - error.code uses UPPER_SNAKE format
//
// Example:
//
//	func CreateNote(w http.ResponseWriter, r *http.Request) error {
//	    var input CreateNoteInput
//	    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
//	        return errors.NewDomain(errors.CodeBadRequest, "invalid request body")
//	    }
//	    note, err := noteService.Create(r.Context(), input)
//	    if err != nil {
//	        return err
//	    }
//	    response.SuccessEnvelopeWithStatus(w, http.StatusCreated, r.Context(), note)
//	    return nil
//	}
func WrapHandler(h HandlerFuncE) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			response.HandleErrorCtx(w, r.Context(), err)
		}
	}
}
