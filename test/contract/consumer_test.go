//go:build contract

package contract

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// MockAuthToken is a valid JWT token structure for testing
	MockAuthToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
)

// TestConsumerHealthEndpoint verifies the health endpoint contract from consumer perspective
func TestConsumerHealthEndpoint(t *testing.T) {
	config := DefaultConfig()

	// Create a new Pact mock server
	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	// Define the expected interaction for the health endpoint
	err = mockProvider.
		AddInteraction().
		UponReceiving("a request to the health endpoint").
		WithRequest("GET", "/healthz").
		WillRespondWith(200, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/json"))
			b.JSONBody(map[string]interface{}{})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			resp, err := http.Get(fmt.Sprintf("http://%s:%d/healthz", config.Host, config.Port))
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected status 200, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "health endpoint contract failed")
}

// TestConsumerReadinessEndpoint verifies the readiness endpoint contract
func TestConsumerReadinessEndpoint(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		UponReceiving("a request to the readiness endpoint").
		WithRequest("GET", "/readyz").
		WillRespondWith(200, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/json"))
			b.JSONBody(map[string]interface{}{})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			resp, err := http.Get(fmt.Sprintf("http://%s:%d/readyz", config.Host, config.Port))
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected status 200, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "readiness endpoint contract failed")
}

// TestConsumerListUsers verifies the list users endpoint contract
func TestConsumerListUsers(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		Given("users exist").
		UponReceiving("a request to list users").
		WithRequest("GET", "/api/v1/users", func(b *consumer.V4RequestBuilder) {
			b.Query("page", matchers.S("1"))
			b.Query("page_size", matchers.S("10"))
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken))
		}).
		WillRespondWith(200, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/json"))
			b.JSONBody(map[string]interface{}{
				"data": matchers.EachLike(map[string]interface{}{
					"id":        matchers.Like("0193e456-7e89-7123-a456-426614174000"), // UUID v7 sample
					"firstName": matchers.Like("John"),
					"lastName":  matchers.Like("Doe"),
					"email":     matchers.Like("user@example.com"),
					"createdAt": matchers.Like("2024-01-01T00:00:00Z"),
					"updatedAt": matchers.Like("2024-01-01T00:00:00Z"),
				}, 1),
				"pagination": map[string]interface{}{
					"page":       matchers.Integer(1),
					"pageSize":   matchers.Integer(10),
					"totalItems": matchers.Integer(1),
					"totalPages": matchers.Integer(1),
				},
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/api/v1/users?page=1&page_size=10", config.Host, config.Port), nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected status 200, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "list users endpoint contract failed")
}

// TestConsumerGetUserByID verifies the get user by ID endpoint contract
func TestConsumerGetUserByID(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		Given("a user exists").
		UponReceiving("a request to get a user by ID").
		WithRequest("GET", "/api/v1/users/0193e456-7e89-7123-a456-426614174000", func(b *consumer.V4RequestBuilder) {
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken))
		}).
		WillRespondWith(200, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/json"))
			b.JSONBody(map[string]interface{}{
				"data": map[string]interface{}{
					"id":        matchers.Like("0193e456-7e89-7123-a456-426614174000"),
					"firstName": matchers.Like("John"),
					"lastName":  matchers.Like("Doe"),
					"email":     matchers.Like("user@example.com"),
					"createdAt": matchers.Like("2024-01-01T00:00:00Z"),
					"updatedAt": matchers.Like("2024-01-01T00:00:00Z"),
				},
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/api/v1/users/0193e456-7e89-7123-a456-426614174000", config.Host, config.Port), nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected status 200, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "get user by ID endpoint contract failed")
}

// TestConsumerGetUserNotFound verifies the 404 error response contract
func TestConsumerGetUserNotFound(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		UponReceiving("a request to get a non-existent user").
		WithRequest("GET", "/api/v1/users/0193e456-7e89-7123-a456-426614174999", func(b *consumer.V4RequestBuilder) {
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken+"-vp-admin"))
		}).
		WillRespondWith(404, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/problem+json"))
			b.JSONBody(map[string]interface{}{
				"type":   matchers.Like("https://api.example.com/problems/not-found"),
				"title":  "User Not Found",
				"status": 404,
				"detail": matchers.Like("User not found"),
				"code":   "USR-001",
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/api/v1/users/0193e456-7e89-7123-a456-426614174999", config.Host, config.Port), nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusNotFound {
				return fmt.Errorf("expected status 404, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "get user not found contract failed")
}

// TestConsumerRateLimitExceeded verifies the 429 rate limit response contract
func TestConsumerRateLimitExceeded(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		Given("rate limit exceeded").
		UponReceiving("a request when rate limit is EXHAUSTED").
		WithRequest("GET", "/api/v1/users", func(b *consumer.V4RequestBuilder) {
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken))
		}).
		WillRespondWith(429, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.S("application/problem+json"))
			b.Header("X-RateLimit-Limit", matchers.Integer(100))
			b.Header("X-RateLimit-Remaining", matchers.Integer(0))
			b.Header("Retry-After", matchers.Integer(60))
			b.JSONBody(map[string]interface{}{
				"type":   matchers.Like("https://api.example.com/problems/rate-limit-exceeded"),
				"title":  "Rate Limit Exceeded",
				"status": 429,
				"detail": matchers.Like("Rate limit exceeded"),
				"code":   "RATE-001",
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/api/v1/users", config.Host, config.Port), nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusTooManyRequests {
				return fmt.Errorf("expected status 429, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "rate limit exceeded contract failed")
}

// =============================================================================
// Story 5.2: Additional Consumer Contract Tests
// =============================================================================

// TestConsumerCreateUser verifies the POST /api/v1/users endpoint contract
func TestConsumerCreateUser(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		Given("valid user data").
		UponReceiving("a request to create a user").
		WithRequest("POST", "/api/v1/users", func(b *consumer.V4RequestBuilder) {
			b.Header("Content-Type", matchers.S("application/json"))
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken))
			b.Header("Idempotency-Key", matchers.UUID())
			b.JSONBody(map[string]interface{}{
				"firstName": matchers.Like("Jane"),
				"lastName":  matchers.Like("Smith"),
				"email":     matchers.Like("jane.smith@example.com"),
			})
		}).
		WillRespondWith(201, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/json"))
			b.JSONBody(map[string]interface{}{
				"data": map[string]interface{}{
					"id":        matchers.UUID(),
					"firstName": matchers.Like("Jane"),
					"lastName":  matchers.Like("Smith"),
					"email":     matchers.Like("jane.smith@example.com"),
					"createdAt": matchers.Like("2024-01-01T00:00:00Z"),
					"updatedAt": matchers.Like("2024-01-01T00:00:00Z"),
				},
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			reqBody := `{"firstName":"Jane","lastName":"Smith","email":"jane.smith@example.com"}`
			req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/api/v1/users", config.Host, config.Port), strings.NewReader(reqBody))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)
			req.Header.Set("Idempotency-Key", "550e8400-e29b-41d4-a716-446655440000")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				return fmt.Errorf("expected status 201, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "create user endpoint contract failed")
}

// TestConsumerUpdateUser verifies the PUT /api/v1/users/{id} endpoint contract
func TestConsumerUpdateUser(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		Given("user exists for update").
		UponReceiving("a request to update a user").
		WithRequest("PUT", "/api/v1/users/0193e456-7e89-7123-a456-426614174000", func(b *consumer.V4RequestBuilder) {
			b.Header("Content-Type", matchers.S("application/json"))
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken))
			b.JSONBody(map[string]interface{}{
				"firstName": matchers.Like("John Updated"),
				"lastName":  matchers.Like("Doe Updated"),
			})
		}).
		WillRespondWith(200, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/json"))
			b.JSONBody(map[string]interface{}{
				"data": map[string]interface{}{
					"id":        matchers.Like("0193e456-7e89-7123-a456-426614174000"),
					"firstName": matchers.Like("John Updated"),
					"lastName":  matchers.Like("Doe Updated"),
					"email":     matchers.Like("user@example.com"),
					"createdAt": matchers.Like("2024-01-01T00:00:00Z"),
					"updatedAt": matchers.Like("2024-01-01T00:00:00Z"),
				},
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			reqBody := `{"firstName":"John Updated","lastName":"Doe Updated"}`
			req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:%d/api/v1/users/0193e456-7e89-7123-a456-426614174000", config.Host, config.Port), strings.NewReader(reqBody))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("expected status 200, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "update user endpoint contract failed")
}

// TestConsumerUpdateUserNotFound verifies the 404 response for update
func TestConsumerUpdateUserNotFound(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		UponReceiving("a request to update a non-existent user").
		WithRequest("PUT", "/api/v1/users/0193e456-7e89-7123-a456-426614174888", func(b *consumer.V4RequestBuilder) {
			b.Header("Content-Type", matchers.S("application/json"))
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken))
			b.JSONBody(map[string]interface{}{
				"firstName": matchers.Like("John"),
			})
		}).
		WillRespondWith(404, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/problem+json"))
			b.JSONBody(map[string]interface{}{
				"type":   matchers.Like("https://api.example.com/problems/not-found"),
				"title":  "User Not Found",
				"status": 404,
				"detail": matchers.Like("User not found"),
				"code":   "USR-001",
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			reqBody := `{"firstName":"John"}`
			req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:%d/api/v1/users/0193e456-7e89-7123-a456-426614174888", config.Host, config.Port), strings.NewReader(reqBody))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusNotFound {
				return fmt.Errorf("expected status 404, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "update user not found contract failed")
}

// TestConsumerDeleteUser verifies the DELETE /api/v1/users/{id} endpoint contract
func TestConsumerDeleteUser(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		Given("user exists for deletion").
		UponReceiving("a request to delete a user").
		WithRequest("DELETE", "/api/v1/users/0193e456-7e89-7123-a456-426614174001", func(b *consumer.V4RequestBuilder) {
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken))
		}).
		WillRespondWith(204, func(b *consumer.V4ResponseBuilder) {
			// 204 No Content - empty response body
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			req, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s:%d/api/v1/users/0193e456-7e89-7123-a456-426614174001", config.Host, config.Port), nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusNoContent {
				return fmt.Errorf("expected status 204, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "delete user endpoint contract failed")
}

// TestConsumerDeleteUserNotFound verifies the 404 response for delete
func TestConsumerDeleteUserNotFound(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		UponReceiving("a request to delete a non-existent user").
		WithRequest("DELETE", "/api/v1/users/0193e456-7e89-7123-a456-426614174777", func(b *consumer.V4RequestBuilder) {
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken))
		}).
		WillRespondWith(404, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/problem+json"))
			b.JSONBody(map[string]interface{}{
				"type":   matchers.Like("https://api.example.com/problems/not-found"),
				"title":  "User Not Found",
				"status": 404,
				"detail": matchers.Like("User not found"),
				"code":   "USR-001",
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			req, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s:%d/api/v1/users/0193e456-7e89-7123-a456-426614174777", config.Host, config.Port), nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusNotFound {
				return fmt.Errorf("expected status 404, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "delete user not found contract failed")
}

// TestConsumerCreateUserValidationError verifies 400 validation error response
func TestConsumerCreateUserValidationError(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		UponReceiving("a request to create user with invalid data").
		WithRequest("POST", "/api/v1/users", func(b *consumer.V4RequestBuilder) {
			b.Header("Content-Type", matchers.S("application/json"))
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken))
			b.JSONBody(map[string]interface{}{
				"firstName": "",              // Invalid: empty
				"lastName":  "Smith",         // Valid
				"email":     "invalid-email", // Invalid: not an email format
			})
		}).
		WillRespondWith(400, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/problem+json"))
			b.JSONBody(map[string]interface{}{
				"type":   matchers.Like("https://api.example.com/problems/validation-error"),
				"title":  "Validation Failed",
				"status": 400,
				"detail": matchers.Like("One or more fields failed validation"),
				"code":   "VAL-000",
				"errors": matchers.EachLike(map[string]interface{}{
					"field":   matchers.Like("firstName"),
					"message": matchers.Like("is required"),
					"code":    matchers.Like("VAL-001"),
				}, 1),
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			reqBody := `{"firstName":"","lastName":"Smith","email":"invalid-email"}`
			req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/api/v1/users", config.Host, config.Port), strings.NewReader(reqBody))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusBadRequest {
				return fmt.Errorf("expected status 400, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "create user validation error contract failed")
}

// TestConsumerUnauthorizedRequest verifies 401 authentication error response
func TestConsumerUnauthorizedRequest(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		UponReceiving("a request without valid authentication").
		WithRequest("GET", "/api/v1/users", func(b *consumer.V4RequestBuilder) {
			// No Authorization header
		}).
		WillRespondWith(401, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/problem+json"))
			b.JSONBody(map[string]interface{}{
				"type":   matchers.Like("https://api.example.com/problems/unauthorized"),
				"title":  "Unauthorized",
				"status": 401,
				"detail": matchers.Like("Missing or invalid authentication token"),
				"code":   "AUTH-001",
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/api/v1/users", config.Host, config.Port), nil)
			if err != nil {
				return err
			}
			// Intentionally NOT setting Authorization header

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusUnauthorized {
				return fmt.Errorf("expected status 401, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "unauthorized request contract failed")
}

// TestConsumerInternalServerError verifies 500 server error response
func TestConsumerInternalServerError(t *testing.T) {
	config := DefaultConfig()

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: config.Consumer,
		Provider: config.Provider,
		PactDir:  config.PactDir,
	})
	require.NoError(t, err, "failed to create mock provider")

	err = mockProvider.
		AddInteraction().
		Given("server error occurs").
		UponReceiving("a request that causes a server error").
		WithRequest("GET", "/api/v1/users/trigger-error-500", func(b *consumer.V4RequestBuilder) {
			b.Header("Authorization", matchers.Like("Bearer "+MockAuthToken))
		}).
		WillRespondWith(500, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", matchers.Like("application/problem+json"))
			b.JSONBody(map[string]interface{}{
				"type":       matchers.Like("https://api.example.com/problems/internal-error"),
				"title":      "Internal Server Error",
				"status":     500,
				"detail":     matchers.Like("An unexpected error occurred"),
				"code":       "SYS-001",
				"request_id": matchers.UUID(),
				"trace_id":   matchers.Like("trace-id-placeholder"),
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/api/v1/users/trigger-error-500", config.Host, config.Port), nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+MockAuthToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusInternalServerError {
				return fmt.Errorf("expected status 500, got %d", resp.StatusCode)
			}

			return nil
		})

	assert.NoError(t, err, "internal server error contract failed")
}
