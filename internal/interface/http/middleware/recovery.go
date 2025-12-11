package middleware

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// ErrorResponse represents a generic error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Recovery middleware recovers from panics and returns 500.
// The panic is logged with stack trace for debugging.
// Response body contains only generic error message (no stack trace).
func Recovery(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log panic with stack trace
					logger.Error("panic recovered",
						zap.Any("error", err),
						zap.String("request_id", GetRequestID(r.Context())),
						zap.Stack("stacktrace"),
					)

					// Return generic 500 error - no stack trace in response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					if encErr := json.NewEncoder(w).Encode(ErrorResponse{
						Error: "internal server error",
					}); encErr != nil {
						logger.Error("failed to encode error response", zap.Error(encErr))
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
