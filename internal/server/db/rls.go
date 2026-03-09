package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// RLS (Row Level Security) SQL statements for PostgreSQL multi-tenancy.
// These are applied after Ent migrations.

const rlsMigrationSQL = `
-- Enable RLS on organization-scoped tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE dashboards ENABLE ROW LEVEL SECURITY;
ALTER TABLE dashboard_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE data_sources ENABLE ROW LEVEL SECURITY;
ALTER TABLE saved_queries ENABLE ROW LEVEL SECURITY;
ALTER TABLE integrations ENABLE ROW LEVEL SECURITY;
ALTER TABLE alerts ENABLE ROW LEVEL SECURITY;

-- Force RLS for table owners (important for superusers)
ALTER TABLE users FORCE ROW LEVEL SECURITY;
ALTER TABLE dashboards FORCE ROW LEVEL SECURITY;
ALTER TABLE dashboard_versions FORCE ROW LEVEL SECURITY;
ALTER TABLE data_sources FORCE ROW LEVEL SECURITY;
ALTER TABLE saved_queries FORCE ROW LEVEL SECURITY;
ALTER TABLE integrations FORCE ROW LEVEL SECURITY;
ALTER TABLE alerts FORCE ROW LEVEL SECURITY;

-- Drop existing policies if they exist (idempotent)
DROP POLICY IF EXISTS org_isolation_users ON users;
DROP POLICY IF EXISTS org_isolation_dashboards ON dashboards;
DROP POLICY IF EXISTS org_isolation_dashboard_versions ON dashboard_versions;
DROP POLICY IF EXISTS org_isolation_data_sources ON data_sources;
DROP POLICY IF EXISTS org_isolation_saved_queries ON saved_queries;
DROP POLICY IF EXISTS org_isolation_integrations ON integrations;
DROP POLICY IF EXISTS org_isolation_alerts ON alerts;

-- Create RLS policies using app.current_organization setting
-- Users: filter by memberships table
CREATE POLICY org_isolation_users ON users
    USING (id IN (
        SELECT user_id FROM memberships
        WHERE organization_id = current_setting('app.current_organization', true)::uuid
    ));

-- Dashboards: filter by organization edge
CREATE POLICY org_isolation_dashboards ON dashboards
    USING (organization_dashboards = current_setting('app.current_organization', true)::uuid);

-- Dashboard versions: filter through dashboard
CREATE POLICY org_isolation_dashboard_versions ON dashboard_versions
    USING (dashboard_dashboard_versions IN (
        SELECT id FROM dashboards
        WHERE organization_dashboards = current_setting('app.current_organization', true)::uuid
    ));

-- Data sources: filter by organization edge
CREATE POLICY org_isolation_data_sources ON data_sources
    USING (organization_datasources = current_setting('app.current_organization', true)::uuid);

-- Saved queries: filter by organization edge
CREATE POLICY org_isolation_saved_queries ON saved_queries
    USING (organization_queries = current_setting('app.current_organization', true)::uuid);

-- Integrations: filter by organization edge
CREATE POLICY org_isolation_integrations ON integrations
    USING (organization_integrations = current_setting('app.current_organization', true)::uuid);

-- Alerts: filter by organization edge
CREATE POLICY org_isolation_alerts ON alerts
    USING (organization_alerts = current_setting('app.current_organization', true)::uuid);

-- Allow bypass for service account (optional, for admin operations)
-- This requires a specific role, not used by default
CREATE POLICY service_bypass_users ON users
    USING (current_setting('app.bypass_rls', true) = 'true');
CREATE POLICY service_bypass_dashboards ON dashboards
    USING (current_setting('app.bypass_rls', true) = 'true');
CREATE POLICY service_bypass_dashboard_versions ON dashboard_versions
    USING (current_setting('app.bypass_rls', true) = 'true');
CREATE POLICY service_bypass_data_sources ON data_sources
    USING (current_setting('app.bypass_rls', true) = 'true');
CREATE POLICY service_bypass_saved_queries ON saved_queries
    USING (current_setting('app.bypass_rls', true) = 'true');
CREATE POLICY service_bypass_integrations ON integrations
    USING (current_setting('app.bypass_rls', true) = 'true');
CREATE POLICY service_bypass_alerts ON alerts
    USING (current_setting('app.bypass_rls', true) = 'true');
`

// ApplyRLS applies Row Level Security policies to the database.
// This should be called after Ent migrations.
func ApplyRLS(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, rlsMigrationSQL)
	if err != nil {
		return fmt.Errorf("applying RLS policies: %w", err)
	}
	return nil
}

// SetOrganizationContext sets the current organization for RLS policies.
// This should be called at the start of each request.
func SetOrganizationContext(ctx context.Context, db *sql.DB, orgID uuid.UUID) error {
	_, err := db.ExecContext(ctx, "SET LOCAL app.current_organization = $1", orgID.String())
	if err != nil {
		return fmt.Errorf("setting organization context: %w", err)
	}
	return nil
}

// ClearOrganizationContext clears the organization context.
func ClearOrganizationContext(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "RESET app.current_organization")
	if err != nil {
		return fmt.Errorf("clearing organization context: %w", err)
	}
	return nil
}

// SetBypassRLS enables RLS bypass for admin operations.
// Use with caution - this allows access to all organization data.
func SetBypassRLS(ctx context.Context, db *sql.DB, bypass bool) error {
	val := "false"
	if bypass {
		val = "true"
	}
	_, err := db.ExecContext(ctx, "SET LOCAL app.bypass_rls = $1", val)
	if err != nil {
		return fmt.Errorf("setting RLS bypass: %w", err)
	}
	return nil
}
