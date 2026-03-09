# Multi-Tenancy

Dashforge supports multi-tenant deployments using PostgreSQL Row Level Security (RLS) for data isolation.

## Overview

Multi-tenancy allows multiple organizations to share a single Dashforge instance while keeping their data completely isolated.

```
┌─────────────────────────────────────────────┐
│              Dashforge Server                │
├─────────────────────────────────────────────┤
│  Tenant A        │  Tenant B        │ ...   │
│  ┌────────────┐  │  ┌────────────┐  │       │
│  │ Dashboards │  │  │ Dashboards │  │       │
│  │ Users      │  │  │ Users      │  │       │
│  │ Queries    │  │  │ Queries    │  │       │
│  └────────────┘  │  └────────────┘  │       │
├─────────────────────────────────────────────┤
│         PostgreSQL with RLS                  │
└─────────────────────────────────────────────┘
```

## Enabling Multi-Tenancy

```bash
./dashforge-server serve \
  --database-url "$DATABASE_URL" \
  --auto-migrate \
  --enable-rls
```

This:

1. Creates the tenant table
2. Adds tenant relationships to all entities
3. Applies Row Level Security policies

## How RLS Works

Row Level Security enforces tenant isolation at the database level:

```sql
-- Every query automatically includes tenant filtering
SELECT * FROM dashboards;
-- Actually executes as:
SELECT * FROM dashboards
WHERE tenant_id = current_setting('app.current_tenant')::int;
```

Benefits:

- **Defense in depth**: Even if application code has bugs, database enforces isolation
- **Automatic**: No need to add `WHERE tenant_id = ?` to every query
- **Audit-friendly**: Impossible to accidentally access another tenant's data

## Tenant Context

### Setting Tenant Context

The server sets tenant context from:

1. **JWT token** (primary) - `tid` claim
2. **X-Tenant-ID header** - For service-to-service calls
3. **Subdomain** - `tenant-a.dashforge.example.com`
4. **Query parameter** - `?tenant=tenant-a` (development only)

Priority: JWT > Header > Subdomain > Query param

### Middleware

The tenant middleware extracts and sets tenant context:

```go
// internal/server/middleware/tenant.go
func (m *TenantMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID, err := m.extractTenant(r)
        if err != nil {
            http.Error(w, "Tenant required", http.StatusBadRequest)
            return
        }

        // Set PostgreSQL session variable
        db.SetTenantContext(r.Context(), m.db, tenantID)

        // Add to request context
        ctx := context.WithValue(r.Context(), TenantKey, tenantID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Tenant Schema

```go
// ent/schema/tenant.go
type Tenant struct {
    ent.Schema
}

func (Tenant) Fields() []ent.Field {
    return []ent.Field{
        field.String("slug").Unique().NotEmpty(),
        field.String("name").NotEmpty(),
        field.String("domain").Optional(),
        field.Enum("plan").
            Values("free", "pro", "enterprise").
            Default("free"),
        field.Bool("active").Default(true),
        field.JSON("settings", map[string]any{}).Optional(),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (Tenant) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("users", User.Type),
        edge.To("dashboards", Dashboard.Type),
        edge.To("data_sources", DataSource.Type),
        edge.To("saved_queries", SavedQuery.Type),
    }
}
```

## RLS Policies

The server applies these policies:

```sql
-- Enable RLS on all tenant-scoped tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE dashboards ENABLE ROW LEVEL SECURITY;
ALTER TABLE data_sources ENABLE ROW LEVEL SECURITY;
ALTER TABLE saved_queries ENABLE ROW LEVEL SECURITY;

-- Create isolation policies
CREATE POLICY tenant_isolation_users ON users
    USING (tenant_id = current_setting('app.current_tenant', true)::int);

CREATE POLICY tenant_isolation_dashboards ON dashboards
    USING (tenant_id = current_setting('app.current_tenant', true)::int);

-- Admin bypass (for system operations)
CREATE POLICY admin_bypass ON users
    USING (current_setting('app.is_admin', true)::boolean = true);
```

## Creating Tenants

### Via API

```bash
curl -X POST https://dashforge.example.com/api/v1/admin/tenants \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "acme-corp",
    "name": "Acme Corporation",
    "plan": "pro"
  }'
```

### Via Database

```sql
INSERT INTO tenants (slug, name, plan, active, created_at, updated_at)
VALUES ('acme-corp', 'Acme Corporation', 'pro', true, NOW(), NOW());
```

## User-Tenant Relationship

Users belong to a single tenant:

```json
{
  "id": 1,
  "email": "user@acme.com",
  "name": "John Doe",
  "role": "admin",
  "tenantId": 1
}
```

JWT tokens include the tenant ID:

```json
{
  "uid": 1,
  "email": "user@acme.com",
  "role": "admin",
  "tid": 1
}
```

## Cross-Tenant Operations

Some operations need to work across tenants:

### System Admin

```go
// Bypass RLS for system operations
func (s *Server) listAllTenants(ctx context.Context) ([]*ent.Tenant, error) {
    // Set admin mode
    _, _ = s.db.DB().ExecContext(ctx, "SET LOCAL app.is_admin = true")
    return s.db.Client().Tenant.Query().All(ctx)
}
```

### Tenant Switching

Admins can switch tenant context:

```bash
curl https://dashforge.example.com/api/v1/dashboards \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: 2"
```

## Subdomain Routing

Configure DNS and routing for tenant subdomains:

```
acme.dashforge.example.com    → Tenant: acme
globex.dashforge.example.com  → Tenant: globex
```

### Nginx Configuration

```nginx
server {
    listen 443 ssl;
    server_name ~^(?<tenant>.+)\.dashforge\.example\.com$;

    location / {
        proxy_pass http://dashforge;
        proxy_set_header Host $host;
        proxy_set_header X-Tenant-Slug $tenant;
    }
}
```

## Tenant Plans

Configure features by plan:

```yaml
plans:
  free:
    max_users: 5
    max_dashboards: 10
    features:
      - basic_charts
  pro:
    max_users: 50
    max_dashboards: 100
    features:
      - basic_charts
      - advanced_charts
      - api_access
  enterprise:
    max_users: unlimited
    max_dashboards: unlimited
    features:
      - all
```

## Migration Considerations

### Migrating to Multi-Tenancy

If you're adding multi-tenancy to an existing single-tenant installation:

1. Create a default tenant
2. Assign all existing data to the default tenant
3. Enable RLS

```sql
-- Create default tenant
INSERT INTO tenants (slug, name, plan)
VALUES ('default', 'Default Tenant', 'enterprise');

-- Assign existing users
UPDATE users SET tenant_id = 1 WHERE tenant_id IS NULL;

-- Enable RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
```

### Disabling Multi-Tenancy

To run single-tenant:

```bash
./dashforge-server serve --database-url "$DATABASE_URL"
# Don't use --enable-rls
```

Without RLS, all data is accessible to all authenticated users.
