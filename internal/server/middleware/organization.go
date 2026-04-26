// Package middleware provides HTTP middleware for Dashforge Server.
package middleware

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/plexusone/dashforge/internal/server/db"
)

type contextKey string

const (
	// OrganizationIDKey is the context key for organization ID.
	OrganizationIDKey contextKey = "organization_id"
	// OrganizationSlugKey is the context key for organization slug.
	OrganizationSlugKey contextKey = "organization_slug"
)

// OrganizationMiddleware extracts organization from request and sets RLS context.
type OrganizationMiddleware struct {
	db        *sql.DB
	logger    *slog.Logger
	skipPaths map[string]bool
}

// OrganizationConfig configures the organization middleware.
type OrganizationConfig struct {
	// SkipPaths are paths that don't require organization context.
	SkipPaths []string
	// HeaderName is the header to extract organization from (default: X-Organization-ID).
	HeaderName string
}

// NewOrganizationMiddleware creates organization extraction middleware.
func NewOrganizationMiddleware(database *sql.DB, logger *slog.Logger, cfg OrganizationConfig) *OrganizationMiddleware {
	skipPaths := make(map[string]bool)
	for _, p := range cfg.SkipPaths {
		skipPaths[p] = true
	}

	return &OrganizationMiddleware{
		db:        database,
		logger:    logger,
		skipPaths: skipPaths,
	}
}

// Wrap wraps an http.Handler with organization context extraction.
func (m *OrganizationMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip certain paths (health, public endpoints)
		if m.shouldSkip(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract organization ID from various sources
		orgID, orgSlug := m.extractOrganization(r)

		// Set RLS context in database
		if m.db != nil && orgID != uuid.Nil {
			if err := db.SetOrganizationContext(r.Context(), m.db, orgID); err != nil {
				m.logger.Error("failed to set organization context", "error", err, "organization_id", orgID)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		// Add organization to request context
		ctx := context.WithValue(r.Context(), OrganizationIDKey, orgID)
		ctx = context.WithValue(ctx, OrganizationSlugKey, orgSlug)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *OrganizationMiddleware) shouldSkip(path string) bool {
	// Skip exact matches
	if m.skipPaths[path] {
		return true
	}

	// Skip common public paths
	publicPrefixes := []string{
		"/health",
		"/api/v1/auth/",
		"/viewer/",
	}
	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func (m *OrganizationMiddleware) extractOrganization(r *http.Request) (uuid.UUID, string) {
	// Priority order:
	// 1. X-Organization-ID header (for API clients)
	// 2. JWT claim (from auth middleware)
	// 3. Subdomain (for web app)
	// 4. Query parameter (for development)

	// 1. Header
	if orgHeader := r.Header.Get("X-Organization-ID"); orgHeader != "" {
		id, err := uuid.Parse(orgHeader)
		if err == nil {
			return id, ""
		}
		// If not UUID, treat as slug
		return uuid.Nil, orgHeader
	}

	// 2. From auth context (set by auth middleware)
	if orgID, ok := r.Context().Value(OrganizationIDKey).(uuid.UUID); ok && orgID != uuid.Nil {
		return orgID, ""
	}

	// 3. Subdomain extraction (org.dashforge.io)
	host := r.Host
	if idx := strings.Index(host, ":"); idx > 0 {
		host = host[:idx] // Remove port
	}
	parts := strings.Split(host, ".")
	if len(parts) >= 3 {
		// Assume format: org.domain.tld
		slug := parts[0]
		if slug != "www" && slug != "api" {
			return uuid.Nil, slug
		}
	}

	// 4. Query parameter (development only)
	if orgParam := r.URL.Query().Get("_org"); orgParam != "" {
		id, err := uuid.Parse(orgParam)
		if err == nil {
			return id, ""
		}
		return uuid.Nil, orgParam
	}

	return uuid.Nil, ""
}

// OrganizationIDFromContext retrieves the organization ID from context.
func OrganizationIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(OrganizationIDKey).(uuid.UUID)
	return id
}

// OrganizationSlugFromContext retrieves the organization slug from context.
func OrganizationSlugFromContext(ctx context.Context) string {
	slug, _ := ctx.Value(OrganizationSlugKey).(string)
	return slug
}

// RequireOrganization returns middleware that requires organization context.
func RequireOrganization() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := OrganizationIDFromContext(r.Context())
			orgSlug := OrganizationSlugFromContext(r.Context())

			if orgID == uuid.Nil && orgSlug == "" {
				http.Error(w, "Organization context required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
