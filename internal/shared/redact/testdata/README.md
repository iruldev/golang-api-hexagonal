# Testdata Directory

This directory contains test fixtures for redaction tests.

## Available Fixtures

| File | Description |
|------|-------------|
| `sensitive_log.txt` | Sample log with sensitive data for redaction testing |

## Usage

See `internal/transport/http/handler/testdata/README.md` for usage patterns.

## Template Reference

For creating new test files, copy the template from:
`internal/shared/testutil/template_test.go.example`

See `docs/testing-patterns.md` for complete testing guidelines.
