// Package auth provides authentication for DashForge.
// JWT functionality is provided by SystemForge's session/jwt package.
package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	cfjwt "github.com/grokify/systemforge/session/jwt"
	cfmw "github.com/grokify/systemforge/session/middleware"
)

// Re-export SystemForge JWT types for convenience.
type (
	// JWTService is SystemForge's JWT service.
	JWTService = cfjwt.Service

	// Claims is SystemForge's JWT claims.
	Claims = cfjwt.Claims

	// JWTConfig is SystemForge's JWT config.
	JWTConfig = cfjwt.Config

	// TokenPair is SystemForge's token pair.
	TokenPair = cfjwt.TokenPair
)

// NewJWTService creates a new JWT service.
func NewJWTService(cfg *JWTConfig) (*JWTService, error) {
	return cfjwt.NewService(cfg)
}

// DefaultJWTConfig returns a config with sensible defaults.
var DefaultJWTConfig = cfjwt.DefaultConfig

// ClaimsFromContext retrieves JWT claims from the request context.
func ClaimsFromContext(ctx context.Context) *Claims {
	return cfmw.ClaimsFromContext(ctx)
}

// PrincipalIDFromContext retrieves the principal ID from the request context.
func PrincipalIDFromContext(ctx context.Context) uuid.UUID {
	return cfmw.PrincipalIDFromContext(ctx)
}

// PrincipalTypeFromContext retrieves the principal type from the request context.
func PrincipalTypeFromContext(ctx context.Context) string {
	return cfmw.PrincipalTypeFromContext(ctx)
}

// RequirePlatformAdmin middleware requires platform admin access.
func RequirePlatformAdmin() func(http.Handler) http.Handler {
	return cfmw.RequirePlatformAdmin()
}

// RequireAnyRole middleware checks if the JWT has any of the required roles.
func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return cfmw.RequireAnyRole(roles...)
}

// HTTPAuth returns middleware that validates JWT tokens.
func HTTPAuth(jwtService *JWTService) func(http.Handler) http.Handler {
	return cfmw.HTTPAuth(jwtService)
}

// HTTPAuthOptional returns middleware that validates JWT tokens if present.
func HTTPAuthOptional(jwtService *JWTService) func(http.Handler) http.Handler {
	return cfmw.HTTPAuthOptional(jwtService)
}

// Principal type constants.
const (
	PrincipalTypeHuman       = cfjwt.PrincipalTypeHuman
	PrincipalTypeApplication = cfjwt.PrincipalTypeApplication
	PrincipalTypeAgent       = cfjwt.PrincipalTypeAgent
	PrincipalTypeService     = cfjwt.PrincipalTypeService
)
