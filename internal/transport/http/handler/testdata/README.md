# Testdata Directory

This directory contains test fixtures and sample data for handler tests.

## Available Fixtures

| File | Description |
|------|-------------|
| `valid_create_user.json` | Valid user creation request payload |
| `invalid_create_user.json` | Invalid user creation (missing/invalid fields) |

## Usage

Place JSON fixtures, config files, and other test data in this directory.
Load them in tests using:

```go
func loadFixture(t *testing.T, name string) []byte {
    t.Helper()
    data, err := os.ReadFile(filepath.Join("testdata", name))
    require.NoError(t, err)
    return data
}
```

## Naming Convention

- `valid_{operation}.json` - Valid request/response fixtures
- `invalid_{operation}.json` - Invalid input fixtures for error testing
- `{scenario}.json` - Scenario-specific fixtures

Example:
```
testdata/
├── valid_create_user.json
├── invalid_create_user.json
└── fixtures/
    └── sample_user.json
```

## Template Reference

For creating new test files, copy the template from:
```
internal/shared/testutil/template_test.go.example
```

See `docs/testing-patterns.md` for complete testing guidelines.

## Note

Binary fixtures (images, etc.) should be added to `.gitignore`.
JSON and text fixtures should be version controlled.
