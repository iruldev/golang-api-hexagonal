//go:build contract

package contract

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/stretchr/testify/require"
)

// ProviderTestConfig holds configuration for provider verification
type ProviderTestConfig struct {
	// ProviderBaseURL is the base URL of the running provider service
	ProviderBaseURL string
	// PactURLs are the paths or URLs to pact files to verify
	PactURLs []string
	// DB is the database connection for seeding data
	DB *sql.DB
	// JWTSecret is the secret used to sign tokens
	JWTSecret string
}

// DefaultProviderConfig returns configuration for local provider testing
func DefaultProviderConfig(t *testing.T) ProviderTestConfig {
	baseURL := os.Getenv("PROVIDER_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Find pact files in the pacts directory
	pactDir := getPactDir()
	pactFiles, _ := filepath.Glob(filepath.Join(pactDir, "*.json"))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default to the one used in Makefile
		dbURL = "postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbURL)
	require.NoError(t, err, "failed to open database connection")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Default to the one used in local env (usually empty or specific string)
		// If empty in config, we might need to set it to something locally or rely on what the server uses.
		// For tests we assume the server is running with a known secret or we might fail auth.
		// However, in local dev, typically config.go handles defaults.
		// If server is started via `make run`, it might not have JWT_SECRET set unless .env.local has it.
		// Let's assume a default for testing if not set.
		jwtSecret = "default-secret-for-testing-only-at-least-32-chars"
	}

	return ProviderTestConfig{
		ProviderBaseURL: baseURL,
		PactURLs:        pactFiles,
		DB:              db,
		JWTSecret:       jwtSecret,
	}
}

// TestProviderVerification verifies the provider against consumer contracts
// Note: This test requires the provider service to be running
func TestProviderVerification(t *testing.T) {
	if os.Getenv("PACT_PROVIDER_TEST") != "true" {
		t.Skip("Skipping provider test - set PACT_PROVIDER_TEST=true and ensure provider is running")
	}

	config := DefaultProviderConfig(t)
	defer func() { _ = config.DB.Close() }()

	if len(config.PactURLs) == 0 {
		t.Skip("No pact files found - run consumer tests first to generate contracts")
	}

	// Verify provider is running
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(config.ProviderBaseURL + "/healthz")
	if err != nil {
		t.Skipf("Provider not available at %s: %v", config.ProviderBaseURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Start a reverse proxy to inject headers
	proxyURL, proxyClose := startProxy(t, config.ProviderBaseURL, config.JWTSecret)
	defer proxyClose()

	// Create verifier
	verifier := provider.NewVerifier()

	// Verify against pact files
	err = verifier.VerifyProvider(t, provider.VerifyRequest{
		Provider:        ProviderName,
		ProviderBaseURL: proxyURL,
		PactFiles:       config.PactURLs,

		// Provider state handlers
		StateHandlers: models.StateHandlers{
			"a request to the health endpoint":    stateNoOp,
			"a request to the readiness endpoint": stateNoOp,

			"users exist": func(setup bool, _ models.ProviderState) (models.ProviderStateResponse, error) {
				if setup {
					return stateSeedUser(config.DB)
				}
				return nil, nil // Teardown (could implement cleanup)
			},
			"a user exists": func(setup bool, _ models.ProviderState) (models.ProviderStateResponse, error) {
				if setup {
					return stateSeedUser(config.DB)
				}
				return nil, nil
			},
			"a request to get a non-existent user": stateNoOp, // Handled naturally by empty query/non-matching ID

			"rate limit exceeded": func(setup bool, _ models.ProviderState) (models.ProviderStateResponse, error) {
				if setup {
					return stateExhaustRateLimit(config.ProviderBaseURL)
				}
				return nil, nil
			},
		},
	})

	require.NoError(t, err, "provider verification failed")
}

func startProxy(t *testing.T, target string, jwtSecret string) (string, func()) {
	targetURL, err := url.Parse(target)
	require.NoError(t, err)

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		authHeader := req.Header.Get("Authorization")
		if authHeader != "" {
			role := "user"
			// Check if the mock token indicates admin role
			if strings.Contains(authHeader, "vp-admin") {
				role = "admin"
			}

			token, err := generateValidToken(jwtSecret, role)
			if err == nil {
				req.Header.Set("Authorization", "Bearer "+token)
			}
		}
		// Host header must match target
		req.Host = targetURL.Host
	}

	server := httptest.NewServer(proxy)
	return server.URL, server.Close
}

func stateNoOp(_ bool, _ models.ProviderState) (models.ProviderStateResponse, error) {
	return nil, nil
}

func stateSeedUser(db *sql.DB) (models.ProviderStateResponse, error) {
	fmt.Println("DEBUG: stateSeedUser starting...")
	// Seed a user with the specific ID expected by the contract
	// ID: 0193e456-7e89-7123-a456-426614174000 (UUID v7)
	id := "0193e456-7e89-7123-a456-426614174000"
	firstName := "John"
	lastName := "Doe"
	email := "user@example.com"
	createdAt, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")

	res, err := db.Exec(`
		INSERT INTO users (id, first_name, last_name, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE 
		SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, email = EXCLUDED.email
	`, id, firstName, lastName, email, createdAt, updatedAt)

	if err != nil {
		fmt.Printf("DEBUG: stateSeedUser failed: %v\n", err)
		return nil, fmt.Errorf("failed to seed user: %w", err)
	}
	rows, _ := res.RowsAffected()
	fmt.Printf("DEBUG: stateSeedUser success. Rows affected: %d\n", rows)

	// Verify it exists immediately
	var count int
	_ = db.QueryRow("SELECT count(*) FROM users WHERE id = $1", id).Scan(&count)
	fmt.Printf("DEBUG: Verified user count: %d\n", count)

	return nil, nil
}

func stateExhaustRateLimit(baseURL string) (models.ProviderStateResponse, error) {
	// Send enough requests to exhaust the rate limit
	// Limit is 100 req/min (or sec depending on config)
	// We send 110 requests
	client := &http.Client{Timeout: 1 * time.Second}
	token, _ := generateValidToken("default-secret-for-testing-only-at-least-32-chars", "user") // Use default secret as fallback or pass in config

	var errCount int
	for i := 0; i < 110; i++ {
		req, _ := http.NewRequest("GET", baseURL+"/api/v1/users", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		} else {
			errCount++
		}
	}
	if errCount > 50 {
		return nil, fmt.Errorf("too many errors during rate limit exhaustion: %d", errCount)
	}
	return nil, nil
}

func generateValidToken(secret string, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  "0193e456-7e89-7123-a456-426614174000",
		"role": role,
		"name": "John Doe",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// TestProviderWithBroker verifies provider against contracts from a Pact Broker
// This is the recommended approach for CI/CD pipelines
func TestProviderWithBroker(t *testing.T) {
	brokerURL := os.Getenv("PACT_BROKER_URL")
	if brokerURL == "" {
		t.Skip("PACT_BROKER_URL not set - skipping broker verification")
	}

	brokerToken := os.Getenv("PACT_BROKER_TOKEN")

	config := DefaultProviderConfig(t)
	defer func() { _ = config.DB.Close() }()

	verifier := provider.NewVerifier()

	verifyRequest := provider.VerifyRequest{
		Provider:        ProviderName,
		ProviderBaseURL: config.ProviderBaseURL,

		BrokerURL:   brokerURL,
		BrokerToken: brokerToken,

		// Enable pending pacts - new contracts won't fail verification
		EnablePending: true,

		// Publish verification results to broker
		PublishVerificationResults: true,
		ProviderVersion:            getProviderVersion(),
		ProviderBranch:             os.Getenv("GIT_BRANCH"),

		// State handlers
		StateHandlers: models.StateHandlers{
			"a request to the health endpoint":    stateNoOp,
			"a request to the readiness endpoint": stateNoOp,
			"a request to list users": func(setup bool, _ models.ProviderState) (models.ProviderStateResponse, error) {
				if setup {
					return stateSeedUser(config.DB)
				}
				return nil, nil
			},
			"a request to get a user by ID": func(setup bool, _ models.ProviderState) (models.ProviderStateResponse, error) {
				if setup {
					return stateSeedUser(config.DB)
				}
				return nil, nil
			},
			"a request to get a non-existent user": stateNoOp,
		},

		RequestFilter: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Replace the mock token with a real valid token
				token, err := generateValidToken(config.JWTSecret, "user")
				if err == nil {
					r.Header.Set("Authorization", "Bearer "+token)
				}
				next.ServeHTTP(w, r)
			})
		},
	}

	err := verifier.VerifyProvider(t, verifyRequest)
	require.NoError(t, err, "provider verification against broker failed")
}

// getProviderVersion returns the version identifier for this provider
func getProviderVersion() string {
	// Use git commit SHA if available
	if sha := os.Getenv("GIT_COMMIT"); sha != "" {
		return sha
	}
	if sha := os.Getenv("GITHUB_SHA"); sha != "" {
		return sha
	}
	// Fallback to timestamp
	return fmt.Sprintf("local-%d", time.Now().Unix())
}
