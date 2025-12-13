package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateModuleName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid lowercase",
			input:   "payment",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "user123",
			wantErr: false,
		},
		{
			name:    "valid with hyphen",
			input:   "user-profile",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			input:   "user_profile",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
			errMsg:  "module name is required",
		},
		{
			name:    "starts with number",
			input:   "123user",
			wantErr: true,
			errMsg:  "module name must start with a letter",
		},
		{
			name:    "uppercase letters",
			input:   "Payment",
			wantErr: true,
			errMsg:  "module name must start with a letter",
		},
		{
			name:    "special characters",
			input:   "pay@ment",
			wantErr: true,
			errMsg:  "module name must start with a letter",
		},
		{
			name:    "spaces",
			input:   "pay ment",
			wantErr: true,
			errMsg:  "module name must start with a letter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateModuleName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModuleName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestSingularize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"payments", "payment"},
		{"users", "user"},
		{"notes", "note"},
		{"payment", "payment"},
		{"user", "user"},
		{"a", "a"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := singularize(tt.input)
			if result != tt.expected {
				t.Errorf("singularize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"payment", "Payment"},
		{"user", "User"},
		{"a", "A"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToSnakeCasePlural(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"payment", "payments"},
		{"user", "users"},
		{"user-profile", "user_profiles"},
		{"payments", "payments"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCasePlural(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCasePlural(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewGenerateCmd(t *testing.T) {
	cmd := newGenerateCmd()

	if cmd.Use != "generate" {
		t.Errorf("expected Use to be 'generate', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	// Check subcommands
	subcmds := cmd.Commands()
	if len(subcmds) == 0 {
		t.Error("expected generate command to have subcommands")
	}

	// Verify module subcommand exists
	found := false
	for _, sub := range subcmds {
		if sub.Use == "module <name>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'module' subcommand to be registered")
	}
}

func TestNewGenerateModuleCmd(t *testing.T) {
	cmd := newGenerateModuleCmd()

	if cmd.Use != "module <name>" {
		t.Errorf("expected Use to be 'module <name>', got %q", cmd.Use)
	}

	// Check flags
	entityFlag := cmd.Flags().Lookup("entity")
	if entityFlag == nil {
		t.Error("expected --entity flag to be defined")
	}
	if entityFlag.Shorthand != "e" {
		t.Errorf("expected --entity shorthand to be 'e', got %q", entityFlag.Shorthand)
	}
}

func TestGenerateModuleCmd_MissingArg(t *testing.T) {
	cmd := newGenerateModuleCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing module name argument")
	}
}

func TestGenerateModuleCmd_InvalidName(t *testing.T) {
	cmd := newGenerateModuleCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"Invalid-Name"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid module name")
	}
}

func TestScaffoldModule_Integration(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "bplat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal go.mod file
	goModContent := "module github.com/test/testproject\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(originalWd)

	// Create generate command and execute
	genCmd := newGenerateCmd()
	var out bytes.Buffer
	genCmd.SetOut(&out)
	genCmd.SetErr(&out)
	genCmd.SetArgs([]string{"module", "payment"})

	err = genCmd.Execute()
	if err != nil {
		t.Fatalf("scaffold failed: %v", err)
	}

	// Verify expected files were created
	expectedPaths := []string{
		"internal/domain/payment/entity.go",
		"internal/domain/payment/errors.go",
		"internal/domain/payment/repository.go",
		"internal/domain/payment/entity_test.go",
		"internal/usecase/payment/usecase.go",
		"internal/usecase/payment/usecase_test.go",
		"internal/interface/http/payment/handler.go",
		"internal/interface/http/payment/dto.go",
		"internal/interface/http/payment/handler_test.go",
	}

	for _, path := range expectedPaths {
		fullPath := filepath.Join(tmpDir, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to be created", path)
		}
	}

	// Verify db files exist with timestamp prefix
	dbMigrationDir := filepath.Join(tmpDir, "db", "migrations")
	entries, err := os.ReadDir(dbMigrationDir)
	if err != nil {
		t.Errorf("failed to read migrations dir: %v", err)
	} else {
		// Should have 2 migration files (up and down)
		if len(entries) != 2 {
			t.Errorf("expected 2 migration files, got %d", len(entries))
		}
	}

	// Verify sqlc queries file
	queriesPath := filepath.Join(tmpDir, "db", "queries", "payment.sql")
	if _, err := os.Stat(queriesPath); os.IsNotExist(err) {
		t.Error("expected sqlc queries file to be created")
	}

	// Verify success message
	output := out.String()
	if !contains(output, "Successfully generated module") {
		t.Errorf("expected success message in output, got: %s", output)
	}

	// REVIEW FIX: Verify template content substitution (not just file existence)
	entityPath := filepath.Join(tmpDir, "internal", "domain", "payment", "entity.go")
	entityContent, err := os.ReadFile(entityPath)
	if err != nil {
		t.Fatalf("failed to read entity file: %v", err)
	}

	// Verify template variables are correctly substituted
	contentChecks := []struct {
		expected string
		desc     string
	}{
		{"package payment", "package name should be module name"},
		{"type Payment struct", "entity struct should use PascalCase entity name"},
		{"func NewPayment()", "constructor should use PascalCase entity name"},
		{"func (e *Payment) Validate()", "Validate method should use entity name"},
	}

	for _, check := range contentChecks {
		if !contains(string(entityContent), check.expected) {
			t.Errorf("entity.go: expected %q (%s)", check.expected, check.desc)
		}
	}

	// Verify repository template substitution
	repoPath := filepath.Join(tmpDir, "internal", "domain", "payment", "repository.go")
	repoContent, err := os.ReadFile(repoPath)
	if err != nil {
		t.Fatalf("failed to read repository file: %v", err)
	}

	if !contains(string(repoContent), "type Repository interface") {
		t.Error("repository.go: expected Repository interface definition")
	}
	if !contains(string(repoContent), "*Payment") {
		t.Error("repository.go: expected Payment type reference")
	}

	// Verify sqlc queries template substitution
	queriesContent, err := os.ReadFile(queriesPath)
	if err != nil {
		t.Fatalf("failed to read queries file: %v", err)
	}

	sqlChecks := []struct {
		expected string
		desc     string
	}{
		{"-- name: CreatePayment", "create query should use entity name"},
		{"-- name: GetPaymentByID", "get query should use entity name"},
		{"payments", "table name should be plural snake_case"},
	}

	for _, check := range sqlChecks {
		if !contains(string(queriesContent), check.expected) {
			t.Errorf("queries.sql: expected %q (%s)", check.expected, check.desc)
		}
	}
}

func TestScaffoldModule_ExistingModule(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "bplat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod and existing module directory
	goModContent := "module github.com/test/testproject\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create existing module directory
	existingDir := filepath.Join(tmpDir, "internal", "domain", "payment")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatalf("failed to create existing dir: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(originalWd)

	// Try to generate - should fail
	genCmd := newGenerateCmd()
	var out bytes.Buffer
	genCmd.SetOut(&out)
	genCmd.SetErr(&out)
	genCmd.SetArgs([]string{"module", "payment"})

	err = genCmd.Execute()
	if err == nil {
		t.Error("expected error for existing module")
	}
	if !contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error message, got: %s", err.Error())
	}
}

func TestScaffoldModule_EntityFlag(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "bplat-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goModContent := "module github.com/test/testproject\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(originalWd)

	// Generate with custom entity name
	genCmd := newGenerateCmd()
	var out bytes.Buffer
	genCmd.SetOut(&out)
	genCmd.SetErr(&out)
	genCmd.SetArgs([]string{"module", "orders", "--entity", "Order"})

	err = genCmd.Execute()
	if err != nil {
		t.Fatalf("scaffold failed: %v", err)
	}

	// Verify entity file contains correct entity name
	entityPath := filepath.Join(tmpDir, "internal", "domain", "orders", "entity.go")
	content, err := os.ReadFile(entityPath)
	if err != nil {
		t.Fatalf("failed to read entity file: %v", err)
	}

	// Should contain Order (the custom entity name)
	if !contains(string(content), "type Order struct") {
		t.Errorf("expected entity file to contain 'type Order struct'")
	}
}

func TestModuleTemplateData(t *testing.T) {
	data := ModuleTemplateData{
		ModuleName:      "payment",
		EntityName:      "Payment",
		EntityNameLower: "payment",
		TableName:       "payments",
		Timestamp:       "20251214021630",
		ModulePath:      "github.com/test/project",
	}

	// Verify all fields are set
	if data.ModuleName == "" {
		t.Error("ModuleName should not be empty")
	}
	if data.EntityName == "" {
		t.Error("EntityName should not be empty")
	}
	if data.TableName == "" {
		t.Error("TableName should not be empty")
	}
	if len(data.Timestamp) != 14 {
		t.Errorf("Timestamp should be 14 chars (YYYYMMDDHHMMSS), got %d", len(data.Timestamp))
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
