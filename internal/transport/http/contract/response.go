package contract

import (
	"encoding/json"
	"net/http"
)

// DataResponse is a generic wrapper for success responses.
type DataResponse[T any] struct {
	Data T `json:"data"`
}

// WriteJSON writes a JSON response with the provided status code.
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
