// Package handlers contains HTTP request handlers for the API.
package handlers

import (
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// HealthData represents the health check data.
type HealthData struct {
	Status string `json:"status"`
}

// HealthHandler returns the health status of the service.
// Response format: {"success": true, "data": {"status": "ok"}}
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response.Success(w, HealthData{Status: "ok"})
}
