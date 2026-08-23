# SpiceDB Migration Tasks

> **Status**: Pending
>
> Add SpiceDB authorization to DashForge for fine-grained dashboard permissions.

## Overview

DashForge currently uses a simple role hierarchy in `auth.go` plus PostgreSQL RLS. This migration adds SpiceDB for:

- Fine-grained dashboard/resource permissions
- Cross-organization sharing capabilities
- Consistent authorization across SystemForge ecosystem

## Current Architecture

**Authentication**: JWT + OAuth (GitHub, Google, CoreControl)
**Authorization**: Simple role hierarchy (admin > editor > viewer)
**Data Isolation**: PostgreSQL RLS via `internal/server/db/rls.go`

**Roles in Membership table**: owner, admin, editor, viewer
**Permissions field**: JSON array (currently unused)

## Tasks

### Phase 1: Add SpiceDB Package

- [x] Create `internal/authz/` package
- [x] Add `github.com/grokify/systemforge/authz/spicedb` import
- [x] Update go.mod with SpiceDB dependencies (SystemForge v0.2.0)
- [x] Run `go mod tidy`

### Phase 2: Define SpiceDB Schema

- [x] Create `internal/authz/schema.go`:
  ```zed
  definition dashboard {
      relation org: organization
      relation owner: principal
      relation editor: principal
      relation viewer: principal

      permission manage = owner + org->admin
      permission edit = manage + editor + org->editor
      permission view = edit + viewer + org->viewer
      permission delete = owner + org->admin
  }

  definition data_source {
      relation org: organization
      relation owner: principal

      permission manage = owner + org->admin
      permission use = manage + org->editor
      permission view = use + org->viewer
  }

  definition saved_query {
      relation org: organization
      relation owner: principal
      relation shared_with: principal

      permission manage = owner
      permission execute = manage + shared_with + org->editor
      permission view = execute + org->viewer
  }

  definition alert {
      relation org: organization
      relation owner: principal

      permission manage = owner + org->admin
      permission view = manage + org->member
  }
  ```

### Phase 3: Create AuthZ Service

- [x] Create `internal/authz/service.go`:
  - SpiceDB client wrapper
  - Permission check methods
  - Role lookup from database
- [x] Service supports both simple and SpiceDB modes
- [ ] Create `internal/authz/syncer.go` (optional - using SystemForge syncer)

### Phase 4: Replace Role Hierarchy

- [ ] Update `internal/server/auth/auth.go`:
  - Remove simple role hierarchy (lines 133-140)
  - Add SpiceDB permission checks
  - Keep `RequireRole()` middleware signature for compatibility
  - Implement using SpiceDB internally

### Phase 5: Wire SpiceDB to Server

- [x] Update `internal/server/server.go`:
  - Initialize SpiceDB client
  - Write schema on startup
  - Add `authzService` field to Server struct
- [x] Add configuration for SpiceDB mode/endpoint (`internal/server/config/config.go`)
- [x] Keep RLS for baseline org isolation

### Phase 6: Update API Handlers

- [ ] `internal/server/api/api.go` - Dashboard operations
- [ ] `internal/server/api/datasource.go` - Data source access
- [ ] `internal/server/api/integration.go` - Integration permissions
- [ ] `internal/server/api/alert.go` - Alert management

### Phase 7: Sync Relationships

- [ ] On membership change → sync to SpiceDB
- [ ] On dashboard creation → sync owner relationship
- [ ] On data source creation → sync owner relationship
- [ ] On sharing → create viewer/editor relationships

### Phase 8: Enable Fine-Grained Permissions

- [ ] Implement dashboard sharing (user-to-dashboard relationships)
- [ ] Implement query sharing
- [ ] Use `membership.permissions` JSON for custom permissions
- [ ] Sync custom permissions to SpiceDB

### Phase 9: Migration Script

- [ ] Create migration script:
  - Read existing memberships
  - Read resource ownership
  - Write relationships to SpiceDB
- [ ] Test with staging data

### Phase 10: Testing

- [ ] Add integration tests
- [ ] Test dashboard permissions
- [ ] Test sharing scenarios
- [ ] Test cross-org access (should be denied)

## Verification

```bash
# Build and test
go build ./...
go test ./...

# Manual verification
# 1. Create org, add members with different roles
# 2. Create dashboard, verify role-based access
# 3. Share dashboard with specific user
# 4. Test cross-org isolation
```

## Notes

- Keep PostgreSQL RLS for org isolation
- SpiceDB for fine-grained resource permissions
- Enables features: dashboard sharing, query sharing, cross-user collaboration

## Dependencies

- SystemForge v0.2.x with SpiceDB support
