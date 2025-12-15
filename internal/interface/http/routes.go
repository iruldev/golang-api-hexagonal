// Package http provides HTTP server and routing functionality.
package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/handlers"
)

// RegisterRoutes registers all API routes under the /api/v1 prefix.
//
// # Adding a New Handler
//
// 1. Create handler function in internal/interface/http/handlers/
// 2. Add route registration here using r.Method("/path", handlers.YourHandler)
// 3. Middleware chain (Recovery, RequestID, Otel, Logging) is applied automatically
//
// # Route Prefixing
//
// All routes registered here are automatically prefixed with /api/v1
// Example: r.Get("/users", ...) becomes GET /api/v1/users
//
// # Available HTTP Methods
//
//	r.Get("/path", handler)     // GET request
//	r.Post("/path", handler)    // POST request
//	r.Put("/path", handler)     // PUT request
//	r.Delete("/path", handler)  // DELETE request
//	r.Patch("/path", handler)   // PATCH request
//
// # URL Parameters
//
//	r.Get("/users/{id}", handler)  // Access via chi.URLParam(r, "id")
//
// # Example Usage
//
//	func RegisterRoutes(r chi.Router) {
//	    r.Get("/health", handlers.HealthHandler)
//	    r.Get("/users", handlers.ListUsers)
//	    r.Post("/users", handlers.CreateUser)
//	    r.Get("/users/{id}", handlers.GetUser)
//	}
func RegisterRoutes(r chi.Router) {
	// Health check (Story 3.1)
	r.Get("/health", handlers.HealthHandler)

	// Example handler demonstrating the pattern (Story 2.4 & 3.6)
	r.Get("/example", WrapHandler(handlers.ExampleHandler))

	// ---------------------------------------------------------------------
	// Add new routes below this line
	// ---------------------------------------------------------------------

}
