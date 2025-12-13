package auth

import (
	"testing"
)

func TestRole_String(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want string
	}{
		{
			name: "admin role string",
			role: RoleAdmin,
			want: "admin",
		},
		{
			name: "service role string",
			role: RoleService,
			want: "service",
		},
		{
			name: "user role string",
			role: RoleUser,
			want: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.String(); got != tt.want {
				t.Errorf("Role.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{
			name: "admin is valid",
			role: RoleAdmin,
			want: true,
		},
		{
			name: "service is valid",
			role: RoleService,
			want: true,
		},
		{
			name: "user is valid",
			role: RoleUser,
			want: true,
		},
		{
			name: "unknown role is invalid",
			role: Role("unknown"),
			want: false,
		},
		{
			name: "empty role is invalid",
			role: Role(""),
			want: false,
		},
		{
			name: "case sensitive - Admin is invalid",
			role: Role("Admin"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.want {
				t.Errorf("Role.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermission_String(t *testing.T) {
	tests := []struct {
		name       string
		permission Permission
		want       string
	}{
		{
			name:       "note:create permission string",
			permission: PermNoteCreate,
			want:       "note:create",
		},
		{
			name:       "note:read permission string",
			permission: PermNoteRead,
			want:       "note:read",
		},
		{
			name:       "note:update permission string",
			permission: PermNoteUpdate,
			want:       "note:update",
		},
		{
			name:       "note:delete permission string",
			permission: PermNoteDelete,
			want:       "note:delete",
		},
		{
			name:       "note:list permission string",
			permission: PermNoteList,
			want:       "note:list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.permission.String(); got != tt.want {
				t.Errorf("Permission.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleConstants(t *testing.T) {
	// Verify role constant values match expected strings
	if RoleAdmin != "admin" {
		t.Errorf("RoleAdmin = %v, want admin", RoleAdmin)
	}
	if RoleService != "service" {
		t.Errorf("RoleService = %v, want service", RoleService)
	}
	if RoleUser != "user" {
		t.Errorf("RoleUser = %v, want user", RoleUser)
	}
}

func TestPermissionConstants(t *testing.T) {
	// Verify permission constant values follow resource:action pattern
	expectedPermissions := map[Permission]string{
		PermNoteCreate: "note:create",
		PermNoteRead:   "note:read",
		PermNoteUpdate: "note:update",
		PermNoteDelete: "note:delete",
		PermNoteList:   "note:list",
	}

	for perm, expected := range expectedPermissions {
		if string(perm) != expected {
			t.Errorf("Permission %v = %v, want %v", perm, string(perm), expected)
		}
	}
}

func TestPermission_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		permission Permission
		want       bool
	}{
		{
			name:       "note:create is valid",
			permission: PermNoteCreate,
			want:       true,
		},
		{
			name:       "note:read is valid",
			permission: PermNoteRead,
			want:       true,
		},
		{
			name:       "note:update is valid",
			permission: PermNoteUpdate,
			want:       true,
		},
		{
			name:       "note:delete is valid",
			permission: PermNoteDelete,
			want:       true,
		},
		{
			name:       "note:list is valid",
			permission: PermNoteList,
			want:       true,
		},
		{
			name:       "unknown permission is invalid",
			permission: Permission("unknown:action"),
			want:       false,
		},
		{
			name:       "empty permission is invalid",
			permission: Permission(""),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.permission.IsValid(); got != tt.want {
				t.Errorf("Permission.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
