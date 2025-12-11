package response

import (
	"errors"
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// MapError maps a domain error to HTTP status code and error code.
// Uses errors.Is() for matching wrapped errors.
//
// Returns:
//   - status: HTTP status code
//   - code: Error code string from errors.go (e.g., ERR_NOT_FOUND)
func MapError(err error) (status int, code string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, ErrNotFound
	case errors.Is(err, domain.ErrValidation):
		return http.StatusUnprocessableEntity, ErrValidation
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, ErrUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, ErrForbidden
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, ErrConflict
	case errors.Is(err, domain.ErrInternal):
		return http.StatusInternalServerError, ErrInternalServer
	case errors.Is(err, domain.ErrTimeout):
		return http.StatusGatewayTimeout, ErrTimeout
	default:
		// Unknown errors are treated as internal server errors
		return http.StatusInternalServerError, ErrInternalServer
	}
}

// HandleError writes an error response based on domain error type.
// Automatically maps the domain error to HTTP status and error code.
//
// Example:
//
//	user, err := service.GetUser(id)
//	if err != nil {
//	    response.HandleError(w, err)
//	    return
//	}
//	response.Success(w, user)
func HandleError(w http.ResponseWriter, err error) {
	status, code := MapError(err)
	Error(w, status, code, err.Error())
}
