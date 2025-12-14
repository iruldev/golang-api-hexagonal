package graphql_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/graphql"
	noteuc "github.com/iruldev/golang-api-hexagonal/internal/usecase/note"
)

// TestPlaygroundEnvironmentRestriction verifies that the GraphQL Playground
// is only enabled in development environments (AC 1 & 2 of Story 12.4).
func TestPlaygroundEnvironmentRestriction(t *testing.T) {
	tests := []struct {
		name           string
		env            string
		expectedStatus int
		expectEnabled  bool
	}{
		{
			name:           "Development environment enables playground",
			env:            config.EnvDevelopment,
			expectedStatus: http.StatusOK,
			expectEnabled:  true,
		},
		{
			name:           "Local environment enables playground",
			env:            config.EnvLocal,
			expectedStatus: http.StatusOK,
			expectEnabled:  true,
		},
		{
			name:           "Staging environment disables playground",
			env:            config.EnvStaging,
			expectedStatus: http.StatusNotFound,
			expectEnabled:  false,
		},
		{
			name:           "Production environment disables playground",
			env:            config.EnvProduction,
			expectedStatus: http.StatusNotFound,
			expectEnabled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Create app config with the specified environment
			appCfg := config.AppConfig{
				Name:     "test-app",
				Env:      tt.env,
				HTTPPort: 8080,
			}

			// Create router with environment-gated playground
			router := chi.NewRouter()

			// Register GraphQL endpoint (always available)
			mockRepo := new(MockRepository)
			usecase := noteuc.NewUsecase(mockRepo, zap.NewNop())
			resolver := &graphql.Resolver{NoteUsecase: usecase}
			srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{
				Resolvers: resolver,
			}))
			router.Handle("/query", srv)

			// Register playground only if development environment
			// This mirrors the logic in cmd/server/main.go
			if appCfg.IsDevelopment() {
				router.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
			}

			// Act: Request the playground endpoint
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/playground", nil)
			router.ServeHTTP(w, r)

			// Assert
			if tt.expectEnabled {
				// Playground should return 200 OK with HTML content
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
			} else {
				// Playground should return 404 Not Found
				assert.Equal(t, http.StatusNotFound, w.Code)
			}
		})
	}
}

// TestPlaygroundQueryEndpointConnectivity verifies the playground points to /query (AC 3).
func TestPlaygroundQueryEndpointConnectivity(t *testing.T) {
	// Arrange: Create a development environment router with playground
	appCfg := config.AppConfig{
		Name:     "test-app",
		Env:      config.EnvDevelopment,
		HTTPPort: 8080,
	}

	router := chi.NewRouter()

	// Register GraphQL endpoint
	mockRepo := new(MockRepository)
	usecase := noteuc.NewUsecase(mockRepo, zap.NewNop())
	resolver := &graphql.Resolver{NoteUsecase: usecase}
	srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{
		Resolvers: resolver,
	}))
	router.Handle("/query", srv)

	// Register playground pointing to /query
	if appCfg.IsDevelopment() {
		router.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
	}

	// Act: Request the playground
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/playground", nil)
	router.ServeHTTP(w, r)

	// Assert: Playground HTML should contain reference to /query endpoint
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "/query", "Playground should be configured to use /query endpoint")
}

// TestIsDevelopmentHelper verifies the IsDevelopment() helper method.
func TestIsDevelopmentHelper(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{config.EnvDevelopment, true},
		{config.EnvLocal, true},
		{config.EnvStaging, false},
		{config.EnvProduction, false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := config.AppConfig{Env: tt.env}
			assert.Equal(t, tt.expected, cfg.IsDevelopment())
		})
	}
}
