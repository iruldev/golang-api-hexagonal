// Package auth provides authentication and authorization types for the application.
// This package defines RBAC (Role-Based Access Control) types including roles
// and permissions that can be used throughout the application.
//
// # Roles
//
// The package defines three standard roles:
//   - RoleAdmin: Full system access for administrators
//   - RoleService: Service-to-service authentication
//   - RoleUser: Standard user access
//
// # Permissions
//
// Permissions follow the resource:action pattern (e.g., "note:create").
// Standard CRUD permissions are defined for the note resource.
//
// # Usage Example
//
// Using roles in middleware:
//
//	r.Group(func(r chi.Router) {
//	    r.Use(middleware.AuthMiddleware(jwtAuth))
//	    r.Use(middleware.RequireRole(string(auth.RoleAdmin)))
//	    r.Delete("/users/{id}", deleteUserHandler)
//	})
//
// Using permissions in middleware:
//
//	r.Group(func(r chi.Router) {
//	    r.Use(middleware.AuthMiddleware(jwtAuth))
//	    r.Use(middleware.RequirePermission(string(auth.PermNoteCreate)))
//	    r.Post("/notes", createNoteHandler)
//	})
//
// Checking roles in handler:
//
//	claims, _ := middleware.FromContext(r.Context())
//	if claims.HasRole(string(auth.RoleAdmin)) {
//	    // Admin-specific logic
//	}
package auth

// Role represents a user role in the system.
// Roles define broad access levels and are used for coarse-grained authorization.
type Role string

// Standard roles for RBAC.
// These roles provide a hierarchy of access levels:
//   - Admin has full system access
//   - Service is for machine-to-machine authentication
//   - User is for standard end-user access
const (
	// RoleAdmin represents full system access for administrators.
	// Users with this role can perform any operation in the system.
	RoleAdmin Role = "admin"

	// RoleService represents service-to-service authentication.
	// Used for internal API calls between microservices.
	RoleService Role = "service"

	// RoleUser represents standard user access.
	// The default role for authenticated end-users.
	RoleUser Role = "user"
)

// String returns the string representation of the role.
func (r Role) String() string {
	return string(r)
}

// IsValid checks if the role is one of the defined standard roles.
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleService, RoleUser:
		return true
	}
	return false
}

// Permission represents a granular permission in the system.
// Permissions follow the resource:action pattern for fine-grained access control.
type Permission string

// Standard permissions for CRUD operations on the note resource.
// Additional permissions can be defined as needed for other resources.
const (
	// PermNoteCreate allows creating new notes.
	PermNoteCreate Permission = "note:create"

	// PermNoteRead allows reading notes.
	PermNoteRead Permission = "note:read"

	// PermNoteUpdate allows updating existing notes.
	PermNoteUpdate Permission = "note:update"

	// PermNoteDelete allows deleting notes.
	PermNoteDelete Permission = "note:delete"

	// PermNoteList allows listing all notes.
	PermNoteList Permission = "note:list"
)

// String returns the string representation of the permission.
func (p Permission) String() string {
	return string(p)
}

// IsValid checks if the permission is one of the defined standard permissions.
func (p Permission) IsValid() bool {
	switch p {
	case PermNoteCreate, PermNoteRead, PermNoteUpdate, PermNoteDelete, PermNoteList:
		return true
	}
	return false
}
