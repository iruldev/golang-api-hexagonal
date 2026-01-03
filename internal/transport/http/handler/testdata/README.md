# Testdata Directory

This directory contains test fixtures and sample data for tests.

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
├── invalid_create_user_missing_email.json
└── fixtures/
    └── sample_user.json
```

## Note

Binary fixtures (images, etc.) should be added to `.gitignore`.
JSON and text fixtures should be version controlled.
