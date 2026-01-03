//go:build contract

package contract

import (
	"fmt"
	"net/http"
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
