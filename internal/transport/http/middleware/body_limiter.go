package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

// BodyLimiter enforces a maximum request body size and returns RFC 7807 with HTTP 413 when exceeded.
// It avoids double-writes from MaxBytesReader and bounds memory to maxBytes+1.
func BodyLimiter(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes <= 0 || r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}

			if r.ContentLength > 0 && r.ContentLength > maxBytes {
				handleTooLarge(w, r)
				return
			}

			limited := http.MaxBytesReader(w, r.Body, maxBytes)
			defer func() { _ = limited.Close() }()

			// Read up to maxBytes+1 to detect overflow before calling handler.
			data, err := io.ReadAll(io.LimitReader(limited, maxBytes+1))
			if err != nil {
				var maxErr *http.MaxBytesError
				if errors.As(err, &maxErr) {
					handleTooLarge(w, r)
					return
				}

				contract.WriteProblemJSON(w, r, &app.AppError{
					Op:      "BodyLimiter",
					Code:    app.CodeInternalError,
					Message: "Failed to read request body",
					Err:     err,
				})
				return
			}

			if int64(len(data)) > maxBytes {
				handleTooLarge(w, r)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(data))
			next.ServeHTTP(w, r)
		})
	}
}

func handleTooLarge(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, r.Body)
	contract.WriteProblemJSON(w, r, &app.AppError{
		Op:      "BodyLimiter",
		Code:    app.CodeRequestTooLarge,
		Message: "Request body too large",
	})
}
