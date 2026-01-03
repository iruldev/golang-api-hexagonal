# Testing Patterns

This document describes the testing patterns, conventions, and best practices used in this project.

## Test File Naming Conventions

Test files follow the Go convention with additional category suffixes for large test suites:

### Standard Pattern
```
{component}_test.go           # Default test file
{component}_{category}_test.go  # Split by category when >500 lines
```

### Categories
| Category | Description | Example |
|----------|-------------|---------|
| `unit` | Basic unit tests for logic | `bulkhead_unit_test.go` |
| `concurrent` | Concurrency/stress tests | `bulkhead_concurrent_test.go` |
| `edge` | Edge cases and error handling | `bulkhead_edge_test.go` |
| `integration` | Integration tests (build tag) | `user_repo_integration_test.go` |
| `bench` | Benchmark tests | `circuit_breaker_bench_test.go` |
| `{functionality}` | Domain-specific | `auth_jwt_test.go`, `auth_claims_test.go` |

### Handler Tests
Handler tests are organized by CRUD operation:
```
internal/transport/http/handler/
├── user_create_test.go      # CreateUser handler tests
├── user_get_test.go         # GetUser handler tests
├── user_list_test.go        # ListUsers handler tests
├── user_context_test.go     # Context propagation tests
└── user_helpers_test.go     # Shared test helpers
```

## Test Categorization Guidelines

### Unit Tests
- Test single function/method behavior
- No external dependencies (mocked)
- Fast execution (<100ms)
- Location: Same package as source

### Integration Tests
- Require external resources (database, network)
- Use build tag: `//go:build integration`
- Run separately: `go test -tags=integration ./...`

### Benchmark Tests
- Measure performance characteristics
- Follow naming: `Benchmark{Function}_{Scenario}`
- Example: `BenchmarkCircuitBreaker_HalfOpen`

## Fixture Organization

### testdata/ Directory Structure
```
internal/{layer}/{package}/
└── testdata/
    ├── valid_request.json
    ├── invalid_request.json
    └── fixtures/
        ├── user.json
        └── error_response.json
```

### Loading Fixtures
```go
func loadFixture(t *testing.T, name string) []byte {
    t.Helper()
    data, err := os.ReadFile(filepath.Join("testdata", name))
    require.NoError(t, err)
    return data
}

func TestUserCreate(t *testing.T) {
    body := loadFixture(t, "valid_create_user.json")
    // ... test code
}
```

## Table-Driven Test Template

```go
func TestFunction_Scenario(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr error
    }{
        {
            name:  "valid input",
            input: validInput,
            want:  expectedOutput,
        },
        {
            name:    "invalid input returns error",
            input:   invalidInput,
            wantErr: ErrExpected,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            
            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Mock Generation Workflow

### Using testify/mock
1. Create mock in test file or `mock_{interface}.go`:
```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id domain.ID) (*domain.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.User), args.Error(1)
}
```

2. Setup expectations in test:
```go
mockRepo := new(MockUserRepository)
mockRepo.On("FindByID", mock.Anything, userID).
    Return(&expectedUser, nil)
```

3. Verify expectations:
```go
mockRepo.AssertExpectations(t)
```

## Test Helper Functions

### Common Helpers Location
- Package-specific: `{package}/helpers_test.go`
- Shared across packages: `internal/shared/testutil/`

### Example Helper File
```go
// helpers_test.go
package handler

const testUserResourcePath = "/api/v1/users"

type testProblemDetail struct {
    Type             string            `json:"type"`
    Title            string            `json:"title"`
    Status           int               `json:"status"`
    Code             string            `json:"code,omitempty"`
    ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
}

func createTestUser() domain.User {
    return domain.User{
        ID:        domain.NewID(),
        Email:     "test@example.com",
        FirstName: "John",
        LastName:  "Doe",
    }
}
```

## Code Coverage

### Running Coverage
```bash
# Generate coverage profile
make coverage

# View HTML report
make coverage-html

# Check coverage percentage
make coverage-report
```

### Coverage Thresholds
- Domain layer: ≥80% (enforced)
- Application layer: ≥80% (enforced)
- Infrastructure layer: ≥60% (recommended)

## Best Practices

### DO
- ✅ Use table-driven tests for multiple scenarios
- ✅ Name tests clearly: `TestFunction_Scenario`
- ✅ Use `t.Helper()` for helper functions
- ✅ Split large test files (>500 lines)
- ✅ Keep shared fixtures in `testdata/`
- ✅ Mock external dependencies

### DON'T
- ❌ Create test files over 500 lines
- ❌ Hard-code test data in multiple places
- ❌ Skip error assertion
- ❌ Use global state between tests
- ❌ Ignore race conditions (`go test -race`)

## Linter Configuration

The `funlen` linter enforces function/file size limits:

```yaml
# .golangci.yml
linters:
  enable:
    - funlen

  settings:
    funlen:
      lines: 100        # Max lines per function
      statements: 60    # Max statements per function
      ignore-comments: true
```

To check function lengths:
```bash
golangci-lint run --enable funlen
```

## Running Tests

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/domain/...

# Run with race detector
go test -race ./...

# Run only unit tests (exclude integration)
go test -tags='!integration' ./...

# Run integration tests
go test -tags=integration ./...
```

## Related Documentation

- [Architecture](../_bmad-output/planning-artifacts/architecture.md) - Naming conventions and project structure
- [Coverage Report](../README.md#testing) - Coverage commands and thresholds
