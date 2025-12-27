//go:build tools
// +build tools

// Package tools pins tool dependencies for reproducible builds.
// The mockgen CLI is installed via: go install go.uber.org/mock/mockgen@latest
package tools

import (
	_ "go.uber.org/mock/gomock"
)
