//go:build tools
// +build tools

// Package tools documents development tool dependencies with pinned versions.
// Install all tools via: make bootstrap
//
// Pinned tool versions (keep in sync with bootstrap target in Makefile):
//   - mockgen: v0.6.0 (go.uber.org/mock/mockgen)
//   - sqlc: v1.28.0 (github.com/sqlc-dev/sqlc/cmd/sqlc)
//   - goose: v3.26.0 (github.com/pressly/goose/v3/cmd/goose)
//   - golangci-lint: v1.64.2 (github.com/golangci/golangci-lint/cmd/golangci-lint)
//
// Note: CLI tools cannot be imported as packages. Use `make bootstrap` to install.
package tools

import (
	// gomock is an importable library used by generated mocks
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/pressly/goose/v3/cmd/goose"
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
	_ "go.uber.org/mock/gomock"
)
