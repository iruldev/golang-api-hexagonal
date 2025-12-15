package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// ExampleNewJWTAuthenticator demonstrates creating a JWT authenticator.
func ExampleNewJWTAuthenticator() {
	// Create authenticator with secret key
	// In production, load from environment variable
	secret := []byte("your-secret-key-at-least-32-bytes!!")

	jwtAuth, err := middleware.NewJWTAuthenticator(secret)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Authenticator created: %T\n", jwtAuth)

	// Output:
	// Authenticator created: *middleware.JWTAuthenticator
}

// ExampleNewJWTAuthenticator_withOptions demonstrates using options.
func ExampleNewJWTAuthenticator_withOptions() {
	secret := []byte("your-secret-key-at-least-32-bytes!!")

	// Configure with issuer and audience validation
	jwtAuth, err := middleware.NewJWTAuthenticator(
		secret,
		middleware.WithIssuer("my-app"),
		middleware.WithAudience("my-api"),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Authenticator with options: %T\n", jwtAuth)

	// Output:
	// Authenticator with options: *middleware.JWTAuthenticator
}

// ExampleJWTAuthenticator_Authenticate demonstrates token validation.
func ExampleJWTAuthenticator_Authenticate() {
	secret := []byte("your-secret-key-at-least-32-bytes!!")
	jwtAuth, _ := middleware.NewJWTAuthenticator(secret)

	// Generate a test token (in production, tokens come from login/auth service)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         "user-123",
		"roles":       []string{"admin"},
		"permissions": []string{"read", "write"},
		"exp":         time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(secret)

	// Create request with token
	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Authenticate
	claims, err := jwtAuth.Authenticate(req)
	if err != nil {
		fmt.Println("Authentication failed:", err)
		return
	}

	fmt.Println("UserID:", claims.UserID)
	fmt.Println("Has admin role:", claims.HasRole("admin"))

	// Output:
	// UserID: user-123
	// Has admin role: true
}

// ExampleAuthMiddleware_withJWT demonstrates using JWT with AuthMiddleware.
func ExampleAuthMiddleware_withJWT() {
	secret := []byte("your-secret-key-at-least-32-bytes!!")
	jwtAuth, _ := middleware.NewJWTAuthenticator(secret)

	// Create a protected handler
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := ctxutil.ClaimsFromContext(r.Context())
		if err != nil {
			http.Error(w, "No claims", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Hello, %s!", claims.UserID)
	})

	// Wrap handler with auth middleware
	handler := middleware.AuthMiddleware(jwtAuth, observability.NewNopLoggerInterface(), false)(protectedHandler)

	// Generate test token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "alice",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(secret)

	// Make authenticated request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Body:", w.Body.String())

	// Output:
	// Status: 200
	// Body: Hello, alice!
}

// ExampleAuthMiddleware_unauthorized demonstrates failed authentication.
func ExampleAuthMiddleware_unauthorized() {
	secret := []byte("your-secret-key-at-least-32-bytes!!")
	jwtAuth, _ := middleware.NewJWTAuthenticator(secret)

	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Protected content")
	})

	handler := middleware.AuthMiddleware(jwtAuth, observability.NewNopLoggerInterface(), false)(protectedHandler)

	// Request without token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)

	// Output:
	// Status: 401
}
