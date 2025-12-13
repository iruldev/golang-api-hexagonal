package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
)

// ExampleNewAPIKeyAuthenticator demonstrates creating an API key authenticator
// with the default configuration.
func ExampleNewAPIKeyAuthenticator() {
	// Create a validator from environment variables
	// Environment: API_KEYS="abc123:svc-payments,xyz789:svc-inventory"
	validator := middleware.NewEnvKeyValidator("API_KEYS")

	// Create authenticator with default header (X-API-Key)
	apiAuth, err := middleware.NewAPIKeyAuthenticator(validator)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Authenticator created: %T\n", apiAuth)
	// Output: Authenticator created: *middleware.APIKeyAuthenticator
}

// ExampleNewAPIKeyAuthenticator_customHeader demonstrates creating an API key
// authenticator with a custom header name.
func ExampleNewAPIKeyAuthenticator_customHeader() {
	validator := &middleware.MapKeyValidator{
		Keys: map[string]*middleware.KeyInfo{
			"service-key": {ServiceID: "my-service"},
		},
	}

	// Create authenticator with custom header name
	apiAuth, err := middleware.NewAPIKeyAuthenticator(
		validator,
		middleware.WithHeaderName("X-Custom-API-Key"),
	)
	if err != nil {
		panic(err)
	}

	// Test authentication with custom header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Custom-API-Key", "service-key")

	claims, err := apiAuth.Authenticate(req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Authenticated: %s\n", claims.UserID)
	// Output: Authenticated: my-service
}

// ExampleAPIKeyAuthenticator_withAuthMiddleware demonstrates using the API key
// authenticator with the AuthMiddleware.
func ExampleAPIKeyAuthenticator_withAuthMiddleware() {
	// Create validator with test keys
	validator := &middleware.MapKeyValidator{
		Keys: map[string]*middleware.KeyInfo{
			"payment-service-key": {
				ServiceID:   "svc-payments",
				Roles:       []string{"service", "payments"},
				Permissions: []string{"process-payments", "view-transactions"},
			},
		},
	}

	apiAuth, err := middleware.NewAPIKeyAuthenticator(validator)
	if err != nil {
		panic(err)
	}

	// Create a protected handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, _ := middleware.FromContext(r.Context())
		fmt.Fprintf(w, "Hello, %s!", claims.UserID)
	})

	// Wrap with AuthMiddleware
	protected := middleware.AuthMiddleware(apiAuth)(handler)

	// Make authenticated request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "payment-service-key")
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
	// Output: Hello, svc-payments!
}

// ExampleEnvKeyValidator demonstrates creating a validator from environment variables.
func ExampleEnvKeyValidator() {
	// Set up environment variable for demo
	os.Setenv("DEMO_API_KEYS", "key1:service-a,key2:service-b")
	defer os.Unsetenv("DEMO_API_KEYS")

	// Create validator
	validator := middleware.NewEnvKeyValidator("DEMO_API_KEYS")

	// Validate a key
	keyInfo, err := validator.Validate(nil, "key1")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Service: %s, Roles: %v\n", keyInfo.ServiceID, keyInfo.Roles)
	// Output: Service: service-a, Roles: [service]
}

// ExampleMapKeyValidator demonstrates using an in-memory validator for testing.
func ExampleMapKeyValidator() {
	validator := &middleware.MapKeyValidator{
		Keys: map[string]*middleware.KeyInfo{
			"test-key-1": {
				ServiceID:   "test-service",
				Roles:       []string{"service", "test"},
				Permissions: []string{"test:read", "test:write"},
				Metadata:    map[string]string{"env": "test"},
			},
		},
	}

	keyInfo, err := validator.Validate(nil, "test-key-1")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Service: %s\n", keyInfo.ServiceID)
	fmt.Printf("Roles: %v\n", keyInfo.Roles)
	// Output:
	// Service: test-service
	// Roles: [service test]
}

// ExampleAPIKeyAuthenticator_chiRouter demonstrates integrating API key
// authentication with a chi router.
func ExampleAPIKeyAuthenticator_chiRouter() {
	validator := &middleware.MapKeyValidator{
		Keys: map[string]*middleware.KeyInfo{
			"internal-svc": {ServiceID: "internal"},
		},
	}

	apiAuth, err := middleware.NewAPIKeyAuthenticator(validator)
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()

	// Public routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Protected routes with API key auth
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(apiAuth))
		r.Get("/api/v1/internal", func(w http.ResponseWriter, r *http.Request) {
			claims, _ := middleware.FromContext(r.Context())
			w.Write([]byte("Hello " + claims.UserID))
		})
	})

	// Test protected route
	req := httptest.NewRequest(http.MethodGet, "/api/v1/internal", nil)
	req.Header.Set("X-API-Key", "internal-svc")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
	// Output: Hello internal
}
