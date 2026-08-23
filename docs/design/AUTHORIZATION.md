# DashForge Authorization Architecture

## Overview

DashForge operates as a **two-sided marketplace** for data dashboards:

1. **Dashboard Publishers** - Create and sell dashboard templates
2. **Dashboard Consumers** - Subscribe to DashForge, build reports, and purchase templates

This document defines the authorization model supporting this marketplace.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           DashForge Platform                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────┐     ┌─────────────────────────────────┐    │
│  │    DASHBOARD PUBLISHERS     │     │     DASHBOARD CONSUMERS         │    │
│  │         (Creators)          │     │       (Organizations)           │    │
│  ├─────────────────────────────┤     ├─────────────────────────────────┤    │
│  │                             │     │                                 │    │
│  │  Publisher Org              │     │  Consumer Org                   │    │
│  │  ├── owner                  │     │  ├── owner                      │    │
│  │  ├── admin                  │     │  ├── admin                      │    │
│  │  ├── creator                │     │  ├── editor                     │    │
│  │  └── reviewer               │     │  └── viewer                     │    │
│  │                             │     │                                 │    │
│  │  Creates:                   │     │  Creates:                       │    │
│  │  • Dashboard Templates      │     │  • Private Dashboards           │    │
│  │  • Data Connectors          │     │  • Data Sources (own data)      │    │
│  │  • Integration Recipes      │     │  • Saved Queries                │    │
│  │                             │     │  • Alerts                       │    │
│  │  Publishes to:              │     │                                 │    │
│  │  • Marketplace              │     │  Purchases from:                │    │
│  │                             │     │  • Marketplace                  │    │
│  └─────────────────────────────┘     └─────────────────────────────────┘    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         MARKETPLACE                                  │    │
│  ├─────────────────────────────────────────────────────────────────────┤    │
│  │                                                                      │    │
│  │  Dashboard Templates          Data Connectors                        │    │
│  │  ├── Sales Analytics          ├── Salesforce                        │    │
│  │  ├── Marketing Report         ├── HubSpot                           │    │
│  │  ├── Financial Overview       ├── Google Analytics                  │    │
│  │  └── DevOps Metrics           └── Stripe                            │    │
│  │                                                                      │    │
│  │  Pricing Models:                                                     │    │
│  │  • Free (with attribution)                                           │    │
│  │  • One-time purchase                                                 │    │
│  │  • Subscription                                                      │    │
│  │  • Per-seat licensing                                                │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                      DATA INTEGRATIONS                               │    │
│  ├─────────────────────────────────────────────────────────────────────┤    │
│  │                                                                      │    │
│  │  Built-in:              Third-Party APIs:        Custom:             │    │
│  │  • PostgreSQL           • REST APIs              • Webhooks          │    │
│  │  • MySQL                • GraphQL                • File uploads      │    │
│  │  • Cube.js              • OAuth-connected        • S3/GCS buckets    │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Entity Model

### Organizations

DashForge supports two organization types:

| Type | Description | Capabilities |
|------|-------------|--------------|
| **Publisher** | Creates and sells dashboard templates | Publish to marketplace, manage listings, view analytics |
| **Consumer** | Subscribes to DashForge for internal use | Create dashboards, connect data, purchase templates |
| **Hybrid** | Both publisher and consumer | Full capabilities |

### Resources

| Resource | Owner | Shareable | Marketplace |
|----------|-------|-----------|-------------|
| Dashboard | Organization | Yes (within org, cross-org via marketplace) | Yes |
| Dashboard Template | Publisher Org | Via marketplace purchase | Yes |
| Data Source | Organization | Within org only | No (credentials) |
| Data Connector | Publisher Org | Via marketplace | Yes |
| Saved Query | User/Org | Within org | No |
| Alert | User/Org | Within org | No |
| Integration | Organization | Within org | No |

## SpiceDB Schema

```zed
// =============================================================================
// PRINCIPALS
// =============================================================================

definition principal {}

// =============================================================================
// PLATFORM LEVEL
// =============================================================================

definition platform {
    relation admin: principal
    relation marketplace_moderator: principal

    permission super_admin = admin
    permission moderate_marketplace = admin + marketplace_moderator
    permission view_analytics = admin
}

// =============================================================================
// ORGANIZATIONS
// =============================================================================

// Publisher organization - creates and sells dashboard templates
definition publisher {
    relation owner: principal
    relation admin: principal
    relation creator: principal      // Can create dashboard templates
    relation reviewer: principal     // Can review before publishing

    // Membership hierarchy
    permission manage = owner + admin
    permission create = manage + creator
    permission review = manage + reviewer

    // Publisher operations
    permission delete = owner
    permission settings = manage
    permission billing = owner + admin
    permission view_revenue = manage

    // Template operations
    permission create_template = create
    permission publish_template = manage  // Only admins can publish
    permission view_analytics = manage

    // Connector operations
    permission create_connector = create
    permission publish_connector = manage
}

// Consumer organization - subscribes to DashForge, builds dashboards
definition organization {
    relation owner: principal
    relation admin: principal
    relation editor: principal
    relation viewer: principal

    // Membership hierarchy
    permission manage = owner + admin
    permission edit = manage + editor
    permission view = edit + viewer

    // Organization operations
    permission delete = owner
    permission settings = manage
    permission billing = owner + admin

    // Resource creation
    permission create_dashboard = edit
    permission create_datasource = manage
    permission create_alert = edit
    permission create_integration = manage

    // Marketplace
    permission purchase_template = manage
    permission view_purchases = view
}

// =============================================================================
// DASHBOARDS (Consumer-owned)
// =============================================================================

definition dashboard {
    relation org: organization
    relation owner: principal
    relation editor: principal
    relation viewer: principal
    relation from_template: dashboard_template  // If created from template

    // Permissions
    permission manage = owner + org->admin
    permission edit = manage + editor + org->editor
    permission view = edit + viewer + org->viewer
    permission delete = manage
    permission share = manage
    permission export = edit
}

definition dashboard_version {
    relation dashboard: dashboard

    permission view = dashboard->view
    permission create = dashboard->edit
    permission restore = dashboard->manage
}

// =============================================================================
// DASHBOARD TEMPLATES (Publisher-owned, sold on marketplace)
// =============================================================================

definition dashboard_template {
    relation publisher: publisher
    relation owner: principal
    relation reviewer: principal
    relation licensed_org: organization      // Orgs that purchased this template

    // Publisher permissions
    permission manage = owner + publisher->manage
    permission edit = manage + publisher->creator
    permission review = publisher->review
    permission publish = publisher->publish_template
    permission view_analytics = manage

    // Access via license
    permission use = licensed_org->edit
    permission view = use + edit
}

// =============================================================================
// DATA SOURCES & CONNECTORS
// =============================================================================

// Data source - org-specific, contains credentials
definition data_source {
    relation org: organization
    relation owner: principal
    relation connector: data_connector  // Optional: based on marketplace connector

    permission manage = owner + org->admin
    permission use = manage + org->editor
    permission view = use + org->viewer
    permission test_connection = manage
}

// Data connector - reusable template for data sources (marketplace)
definition data_connector {
    relation publisher: publisher
    relation owner: principal
    relation licensed_org: organization

    permission manage = owner + publisher->manage
    permission edit = manage + publisher->creator
    permission publish = publisher->publish_connector

    // Licensed orgs can use connector
    permission use = licensed_org->manage
    permission view = use + edit
}

// =============================================================================
// SAVED QUERIES & ALERTS
// =============================================================================

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
    relation dashboard: dashboard
    relation subscribers: principal

    permission manage = owner + org->admin
    permission subscribe = org->viewer
    permission view = manage + subscribers + org->viewer
}

// =============================================================================
// INTEGRATIONS (Notification channels)
// =============================================================================

definition integration {
    relation org: organization
    relation owner: principal

    permission manage = owner + org->admin
    permission use = manage + org->editor
    permission view = use + org->viewer
}

// =============================================================================
// MARKETPLACE
// =============================================================================

definition marketplace_listing {
    relation template: dashboard_template
    relation connector: data_connector
    relation publisher: publisher
    relation moderator: principal

    permission manage = publisher->manage
    permission moderate = moderator + platform->moderate_marketplace
    permission view = publisher->creator  // Anyone can view listings
    permission purchase = organization->purchase_template
}

definition template_license {
    relation template: dashboard_template
    relation organization: organization
    relation purchased_by: principal

    permission view = organization->manage + purchased_by
    permission use = organization->edit
    permission transfer = organization->manage
}

definition connector_license {
    relation connector: data_connector
    relation organization: organization
    relation purchased_by: principal

    permission view = organization->manage + purchased_by
    permission use = organization->manage
}
```

## Role Hierarchies

### Publisher Roles

| Role | Level | Capabilities |
|------|-------|--------------|
| owner | 100 | Full control, billing, delete |
| admin | 80 | Manage members, publish templates |
| creator | 60 | Create templates and connectors |
| reviewer | 40 | Review templates before publishing |

### Consumer (Organization) Roles

| Role | Level | Capabilities |
|------|-------|--------------|
| owner | 100 | Full control, billing, delete |
| admin | 80 | Manage members, data sources, integrations |
| editor | 60 | Create/edit dashboards, queries, alerts |
| viewer | 40 | View dashboards and data |

## Authorization Scenarios

### 1. Publisher Creates and Publishes Template

```go
// 1. Creator makes a dashboard template
authz.Can(ctx, creator, "create_template", publisher) // true

// 2. Creator submits for review
authz.Can(ctx, creator, "edit", template) // true

// 3. Reviewer approves
authz.Can(ctx, reviewer, "review", template) // true

// 4. Admin publishes to marketplace
authz.Can(ctx, admin, "publish", template) // true
authz.Can(ctx, creator, "publish", template) // false - only admins
```

### 2. Organization Purchases Template

```go
// 1. Admin purchases from marketplace
authz.Can(ctx, admin, "purchase_template", org) // true

// 2. License is created, synced to SpiceDB
syncer.WriteRelationship(template, "licensed_org", org)

// 3. Editors can now use the template
authz.Can(ctx, editor, "use", template) // true

// 4. Editor creates dashboard from template
dashboard := createFromTemplate(template)
syncer.WriteRelationship(dashboard, "from_template", template)
```

### 3. Cross-Org Dashboard Sharing (Future)

```go
// Share dashboard via public link (read-only)
authz.Can(ctx, externalUser, "view", publicDashboard) // true (if public)

// Embed in external site
authz.Can(ctx, embedToken, "view", dashboard) // true (if embed enabled)
```

## Relationship Sync Events

| Event | SpiceDB Action |
|-------|----------------|
| Publisher created | Create publisher with owner relation |
| Organization created | Create organization with owner relation |
| Member added/role changed | Update membership relations |
| Template created | Add template with publisher and owner relations |
| Template published | Add marketplace_listing relation |
| License purchased | Add licensed_org relation to template |
| Dashboard created | Add org and owner relations |
| Dashboard created from template | Add from_template relation |
| Data source created | Add org, owner, and connector relations |
| Alert created | Add org, owner, and dashboard relations |

## Data Integration Architecture

### Built-in Data Sources

| Type | Description | Auth Method |
|------|-------------|-------------|
| PostgreSQL | Direct database connection | Connection string |
| MySQL | Direct database connection | Connection string |
| Cube.js | Semantic layer | API key |

### Third-Party API Connectors

| Connector | Auth | Marketplace |
|-----------|------|-------------|
| Salesforce | OAuth 2.0 | Yes |
| HubSpot | OAuth 2.0 | Yes |
| Google Analytics | OAuth 2.0 | Yes |
| Stripe | API Key | Yes |
| Custom REST | API Key / OAuth | Template |
| Custom GraphQL | API Key / Bearer | Template |

### Connector Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Connector  │     │ Data Source │     │  Dashboard  │
│  Template   │────▶│  Instance   │────▶│   Widget    │
│ (Reusable)  │     │ (Org-owned) │     │  (Queries)  │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │
       │                   ├── Credentials (encrypted)
       ├── Schema          ├── Connection config
       ├── Auth flow       └── Refresh settings
       └── Field mappings
```

## Pricing Integration

See [FEAT_MARKETPLACE_PRD.md](./FEAT_MARKETPLACE_PRD.md) for:

- Publisher pricing tiers
- Consumer subscription tiers
- Template pricing models
- Revenue sharing

## Migration Strategy

### Phase 1: Core Authorization
- Update SpiceDB schema with publisher/consumer model
- Implement dual-organization service
- Migrate existing orgs as consumer type

### Phase 2: Marketplace Foundation
- Add dashboard_template entity
- Add marketplace_listing entity
- Implement license management

### Phase 3: Data Connectors
- Add data_connector entity
- Build connector template system
- OAuth integration framework

### Phase 4: Full Marketplace
- Payment integration (Stripe Connect)
- Publisher analytics
- Review/rating system
