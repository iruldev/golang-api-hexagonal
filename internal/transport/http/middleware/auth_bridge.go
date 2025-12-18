// Package middleware provides HTTP middleware for the transport layer.
// This file implements the auth context bridge that converts JWT claims to app.AuthContext.
package middleware

import (
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// AuthContextBridge is middleware that converts JWT claims from context to app.AuthContext.
// This bridges the transport layer (ctxutil.Claims) to the app layer (app.AuthContext).
// It should be applied AFTER JWTAuth middleware on protected routes.
//
// The middleware extracts claims set by JWTAuth and creates an app.AuthContext
// that use cases can access via app.GetAuthContext(ctx).
func AuthContextBridge(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only bridge claims that were validated by JWTAuth.
		if !isClaimsValidated(r.Context()) {
			next.ServeHTTP(w, r)
			return
		}

		claims := ctxutil.GetClaims(r.Context())
		if claims != nil {
			// Convert transport-layer claims to app-layer auth context
			authCtx := &app.AuthContext{
				SubjectID: claims.Subject,             // From jwt.RegisteredClaims
				Role:      NormalizeRole(claims.Role), // Normalize role for consistency
			}
			ctx := app.SetAuthContext(r.Context(), authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		// No claims in context - continue without auth context
		// Use cases will handle missing auth context (fail-closed behavior)
		next.ServeHTTP(w, r)
	})
}
