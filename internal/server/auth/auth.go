// Package auth provides authentication and authorization for Dashforge Server.
package auth

import (
	"context"
	"net/http"
)

// User represents an authenticated user.
type User struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Role     Role     `json:"role"`
	Groups   []string `json:"groups,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Role represents a user's permission level.
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// Provider is the interface for authentication providers.
type Provider interface {
	// Authenticate validates credentials and returns a user.
	Authenticate(ctx context.Context, r *http.Request) (*User, error)

	// Name returns the provider name.
	Name() string
}

// Middleware wraps an http.Handler with authentication.
type Middleware struct {
	provider    Provider
	skipPaths   map[string]bool
	disabled    bool
}

// NewMiddleware creates authentication middleware.
func NewMiddleware(provider Provider, opts ...MiddlewareOption) *Middleware {
	m := &Middleware{
		provider:  provider,
		skipPaths: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// MiddlewareOption configures the auth middleware.
type MiddlewareOption func(*Middleware)

// WithSkipPaths sets paths that don't require authentication.
func WithSkipPaths(paths ...string) MiddlewareOption {
	return func(m *Middleware) {
		for _, p := range paths {
			m.skipPaths[p] = true
		}
	}
}

// WithDisabled disables authentication entirely.
func WithDisabled(disabled bool) MiddlewareOption {
	return func(m *Middleware) {
		m.disabled = disabled
	}
}

// Wrap wraps an http.Handler with authentication.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if disabled
		if m.disabled {
			next.ServeHTTP(w, r)
			return
		}

		// Skip certain paths
		if m.skipPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Authenticate
		user, err := m.provider.Authenticate(r.Context(), r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type contextKey string

const userContextKey contextKey = "user"

// UserFromContext retrieves the user from the request context.
func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userContextKey).(*User)
	return user
}

// RequireRole returns middleware that requires a minimum role.
func RequireRole(minRole Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !hasMinRole(user.Role, minRole) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func hasMinRole(userRole, minRole Role) bool {
	roleOrder := map[Role]int{
		RoleViewer: 1,
		RoleEditor: 2,
		RoleAdmin:  3,
	}
	return roleOrder[userRole] >= roleOrder[minRole]
}
