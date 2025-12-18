package middleware

import (
	"context"
	"strings"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// markClaimsValidatedTestOnly marks claims as validated and sets AuthContext for tests.
// Use only in test code; production must rely on JWTAuth.
func markClaimsValidatedTestOnly(ctx context.Context, claims *ctxutil.Claims) context.Context {
	if claims == nil {
		return ctx
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return ctx
	}

	normalizedRole := NormalizeRole(claims.Role)
	claims.Role = normalizedRole

	ctx = ctxutil.SetClaims(ctx, claims) // role normalized in SetClaims as well
	ctx = setValidatedClaims(ctx)
	ctx = app.SetAuthContext(ctx, &app.AuthContext{
		SubjectID: claims.Subject,
		Role:      normalizedRole,
	})
	return ctx
}
