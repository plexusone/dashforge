# Authorization

Dashforge provides flexible authorization with two modes: simple role hierarchy and SpiceDB for fine-grained control.

## Overview

Dashforge has two distinct organization types with separate role hierarchies:

- **Publishers** - Organizations that create and sell dashboard templates
- **Consumers** - Organizations that use Dashforge for their own dashboards

## Role Hierarchies

### Publisher Roles

For organizations creating and selling templates:

| Role | Level | Permissions |
|------|-------|-------------|
| **Owner** | 100 | Full control, billing, delete organization |
| **Admin** | 80 | Manage members, publish templates, view analytics |
| **Creator** | 60 | Create templates and data connectors |
| **Reviewer** | 40 | Review templates before publishing |

### Consumer Roles

For organizations using Dashforge:

| Role | Level | Permissions |
|------|-------|-------------|
| **Owner** | 100 | Full control, billing, delete organization |
| **Admin** | 80 | Manage members, data sources, integrations |
| **Editor** | 60 | Create/edit dashboards, queries, alerts |
| **Viewer** | 40 | View dashboards and data |

## Authorization Modes

### Simple Mode (Default)

Role-based hierarchy with numeric levels:

```go
// Check if user can perform action
if authz.CheckMinRole(ctx, user, "editor") {
    // User has editor role or higher
}

// Check specific permission
if authz.Can(ctx, user, "dashboard:edit", dashboardID) {
    // User can edit this dashboard
}
```

### SpiceDB Mode

Fine-grained relationship-based authorization:

```yaml
# Enable SpiceDB in config
authorization:
  provider: spicedb
  spicedb:
    endpoint: "localhost:50051"
    preshared_key: "your-key"
```

## Resource Permissions

### Dashboard Permissions

| Permission | Description | Roles |
|------------|-------------|-------|
| `view` | View dashboard | Viewer+ |
| `edit` | Edit dashboard | Editor+ |
| `delete` | Delete dashboard | Admin+ |
| `share` | Share with others | Editor+ |
| `export` | Export dashboard | Viewer+ |

### Data Source Permissions

| Permission | Description | Roles |
|------------|-------------|-------|
| `view` | View connection | Viewer+ |
| `use` | Query data | Editor+ |
| `manage` | Create/edit/delete | Admin+ |

### Template Permissions (Publishers)

| Permission | Description | Roles |
|------------|-------------|-------|
| `create` | Create templates | Creator+ |
| `edit` | Edit templates | Creator+ |
| `publish` | Publish to marketplace | Admin+ |
| `delete` | Delete templates | Admin+ |

## SpiceDB Schema

The full authorization schema defines relationships between resources:

```zed
definition platform {
    relation admin: principal
    relation marketplace_moderator: principal
    permission manage = admin
    permission moderate_marketplace = marketplace_moderator + admin
}

definition publisher {
    relation platform: platform
    relation owner: principal
    relation admin: principal
    relation creator: principal
    relation reviewer: principal

    permission manage = owner
    permission publish = admin + owner
    permission create_template = creator + admin + owner
    permission review = reviewer + creator + admin + owner
}

definition organization {
    relation platform: platform
    relation owner: principal
    relation admin: principal
    relation editor: principal
    relation viewer: principal

    permission manage = owner
    permission admin_access = admin + owner
    permission edit = editor + admin + owner
    permission view = viewer + editor + admin + owner
}

definition dashboard {
    relation organization: organization
    relation owner: principal
    relation editor: principal
    relation viewer: principal

    permission manage = owner + organization->manage
    permission edit = editor + owner + organization->edit
    permission view = viewer + editor + owner + organization->view
    permission share = editor + owner + organization->admin_access
    permission export = view
}
```

## Checking Permissions

### In HTTP Handlers

```go
func (h *Handler) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    dashboardID := chi.URLParam(r, "id")

    // Check permission
    if !h.authz.Can(ctx, "dashboard", dashboardID, "edit") {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // ... update logic
}
```

### Filtering Resources

```go
// Get only dashboards user can view
dashboards, err := h.authz.Filter(ctx, "dashboard", "view", allDashboardIDs)
```

### Bulk Permission Check

```go
// Check multiple permissions at once
results, err := h.authz.CanAll(ctx, []authz.Check{
    {Resource: "dashboard", ID: id1, Permission: "edit"},
    {Resource: "dashboard", ID: id2, Permission: "edit"},
})
```

## Relationship Sync

When identity changes occur, relationships are synced to SpiceDB:

| Event | Action |
|-------|--------|
| User joins org | Add `organization#viewer@user` |
| User role changes | Update relation |
| User leaves org | Remove all org relations |
| Dashboard created | Add `dashboard#owner@user` |
| Dashboard shared | Add viewer/editor relation |

## Platform Admin

Platform admins have cross-organization access:

```go
if authz.IsPlatformAdmin(ctx, user) {
    // Full access to all resources
}
```

## Best Practices

1. **Principle of Least Privilege** - Start with Viewer, elevate as needed
2. **Use Groups** - Manage permissions through organization roles
3. **Audit Access** - Log permission checks for compliance
4. **Test Thoroughly** - Verify permission boundaries in tests
5. **Migrate Carefully** - When switching to SpiceDB, run in audit mode first
