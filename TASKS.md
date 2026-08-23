# Dashforge Roadmap

A phased approach to building a Metabase alternative, starting with static dashboards and evolving to a full analytics platform.

## Repository Structure

```
dashforge/
├── cmd/
│   ├── dashforge/              # Static CLI (minimal deps)
│   │   ├── main.go
│   │   └── cmd/
│   │       ├── root.go
│   │       ├── serve.go        # Local dev server
│   │       ├── validate.go     # Dashboard validation
│   │       └── version.go
│   └── dashforge-server/       # Full server (db, auth)
│       ├── main.go
│       └── cmd/
│           ├── root.go
│           ├── serve.go        # Production server
│           └── version.go
├── dashboardir/                # Core IR types (zero deps)
│   ├── dashboard.go
│   ├── datasource.go
│   ├── transform.go
│   └── widget.go
├── viewer/                     # Embedded HTML/JS viewer
│   ├── viewer.go
│   └── index.html
├── schema/                     # JSON Schema
│   ├── schema.go
│   ├── dashboard.schema.json
│   └── generate/
├── internal/
│   └── server/                 # Server-only code
│       ├── server.go
│       ├── api/                # REST API handlers
│       ├── db/                 # Database interface
│       ├── auth/               # Authentication
│       └── config/             # YAML config loading
└── examples/
```

## Current State

### Static CLI (`dashforge`)

- [x] DashboardIR Go types
- [x] JSON Schema generation from Go types
- [x] Static HTML/JS viewer
- [x] URL and inline data sources
- [x] Widgets: chart, metric, table, text
- [x] Transforms: extract, filter, sort, limit
- [x] ChartIR → ECharts compilation
- [x] Interactive variable filters
  - [x] Select dropdowns
  - [x] Text search with debounce
  - [x] Date and daterange pickers
  - [x] URL param sync (shareable links)
  - [x] Dynamic options from data sources
  - [x] Variable interpolation in transforms (`${var:id}`)
- [x] `dashforge serve` - local dev server
- [x] `dashforge validate` - dashboard validation
- [x] Example compliance dashboard

### Server (`dashforge-server`) - Scaffolded

- [x] CLI structure with Cobra
- [x] Server package with HTTP routing
- [x] API handler scaffolding (dashboard CRUD, query execution)
- [x] Database interface definition
- [x] Auth middleware scaffolding
- [x] YAML config loading
- [ ] PostgreSQL implementation (pgx)
- [ ] Dashboard storage (file + database)
- [ ] User management
- [ ] Query caching

---

## Phase 1: Production-Ready Static Dashboards

**Goal:** Make static dashboards robust enough for production GitHub Pages deployment.

### 1.1 Schema & Validation ✅ Complete

- [x] Generate JSON Schema from Go types
- [x] Create `dashforge validate` CLI command
- [ ] Add schema validation to viewer (runtime)
- [ ] Add schema version migration support

### 1.2 Enhanced Viewer

- [ ] Dark mode toggle (use theme from dashboard)
- [ ] Responsive/mobile layout
- [ ] Loading states and error boundaries per widget
- [ ] Keyboard navigation
- [ ] Print/export to PDF

### 1.3 Interactive Filters ✅ Complete

- [x] Implement variable UI controls (select, text, date, daterange)
- [x] Variable-based data filtering
- [x] URL query param sync for variables
- [ ] Cross-widget filtering (click row to filter other widgets)

### 1.4 Improved Widgets

- [ ] Table: client-side sorting, filtering, search
- [ ] Table: column resizing and reordering
- [ ] Metric: sparkline support
- [ ] Metric: comparison to previous period
- [ ] Text: markdown rendering with variable interpolation

### 1.5 More Chart Types (via ChartIR)

- [ ] Heatmap
- [ ] Treemap
- [ ] Funnel
- [ ] Gauge
- [ ] Radar/spider
- [ ] Candlestick (for time series)

### 1.6 Dashboard Features

- [ ] Tab/page support within dashboard
- [ ] Auto-refresh toggle in viewer
- [ ] Full-screen widget mode
- [ ] Widget drill-down navigation
- [ ] Dashboard embedding (iframe snippet generator)

### 1.7 CLI Tools ✅ Partial

- [x] `dashforge serve` - local dev server with live reload
- [ ] `dashforge build` - bundle dashboard + data for deployment
- [ ] `dashforge new` - scaffold new dashboard from template

---

## Phase 2: Server Mode with Database Support

**Goal:** Add a Go server that can query databases directly.

### 2.1 Server Foundation ✅ Scaffolded

- [x] Go HTTP server (`dashforge-server serve`)
- [x] Serve static dashboards
- [x] REST API scaffolding for dashboard CRUD
- [x] Health check endpoint
- [x] Configuration via YAML/env vars
- [ ] Graceful shutdown
- [ ] Metrics endpoint (Prometheus)

### 2.2 PostgreSQL Data Source

- [x] Database interface defined
- [ ] Connection configuration (with env var secrets)
- [ ] Query execution with parameterized queries (pgx)
- [ ] Connection pooling
- [ ] Query timeout and cancellation
- [ ] SSL/TLS support

### 2.3 Query Caching

- [ ] In-memory cache with TTL
- [ ] Cache invalidation on refresh
- [ ] Optional Redis cache backend
- [ ] Cache statistics endpoint

### 2.4 Additional Databases

- [ ] MySQL/MariaDB
- [ ] SQLite (for local/embedded use)
- [ ] ClickHouse (for analytics)
- [ ] DuckDB (for local analytics)

### 2.5 Query Features

- [ ] Named/saved queries
- [ ] Query parameters from variables
- [ ] Query result transformation
- [ ] Query chaining (one query feeds another)

### 2.6 API Data Sources

- [ ] HTTP/REST API data source
- [ ] GraphQL data source
- [ ] Authentication (API key, OAuth)
- [ ] Response transformation

---

## Phase 3: Visual Query Builder

**Goal:** Allow non-technical users to explore data without SQL.

### 3.1 Schema Introspection

- [ ] Auto-discover tables and columns
- [ ] Column type detection
- [ ] Foreign key relationship detection
- [ ] Schema caching

### 3.2 Query Builder UI

- [ ] Table/view selector
- [ ] Column picker with search
- [ ] Filter builder (WHERE clause)
- [ ] Aggregation selector (GROUP BY)
- [ ] Sort configuration (ORDER BY)
- [ ] Limit/pagination

### 3.3 SQL Editor

- [ ] Monaco editor integration
- [ ] Syntax highlighting
- [ ] Auto-complete for tables/columns
- [ ] Query formatting
- [ ] Query history

### 3.4 Question/Query Management

- [ ] Save queries as "questions"
- [ ] Query versioning
- [ ] Query sharing via link
- [ ] Query collections/folders

---

## Phase 4: Collaboration & Multi-User

**Goal:** Support teams with authentication and permissions.

### 4.1 Authentication ✅ Scaffolded

- [x] Auth middleware interface
- [x] Role-based access (admin, editor, viewer)
- [ ] Local user accounts (email/password)
- [ ] Session management (JWT)
- [ ] Password reset flow
- [ ] API tokens for programmatic access

### 4.2 OAuth/SSO

- [ ] GitHub OAuth
- [ ] Google OAuth
- [ ] OIDC generic provider
- [ ] SAML (enterprise)

### 4.3 Authorization

- [ ] Dashboard-level permissions
- [ ] Data source-level permissions
- [ ] Collection-level permissions
- [ ] Row-level security

### 4.4 Collaboration Features

- [ ] Dashboard collections/folders
- [ ] Dashboard favorites
- [ ] Dashboard search
- [ ] Activity feed / audit log
- [ ] Comments on dashboards

### 4.5 Sharing

- [ ] Public dashboard links
- [ ] Password-protected shares
- [ ] Expiring share links
- [ ] Embed tokens with row-level filtering

---

## Phase 5: Enterprise Features

**Goal:** Features required for enterprise adoption.

### 5.1 Scheduled Reports

- [ ] Cron-based dashboard refresh
- [ ] Email delivery (PDF/PNG snapshots)
- [ ] Slack/Teams integration
- [ ] Webhook delivery

### 5.2 Alerts

- [ ] Threshold-based alerts on metrics
- [ ] Alert conditions (above, below, change %)
- [ ] Alert destinations (email, Slack, webhook)
- [ ] Alert history and acknowledgment

### 5.3 Advanced Security

- [ ] Column-level masking
- [ ] IP allowlisting
- [ ] Audit logging with retention

### 5.4 Performance & Scale

- [ ] Query result pagination
- [ ] Async query execution
- [ ] Query queue management
- [ ] Multi-node deployment
- [ ] Read replicas for queries

### 5.5 Administration

- [ ] Admin dashboard (usage stats)
- [ ] User management UI
- [ ] Database connection management UI
- [ ] System health monitoring
- [ ] Backup/restore

---

## Phase 6: Ecosystem & Extensibility

**Goal:** Enable community extensions and integrations.

### 6.1 Plugin System

- [ ] Custom visualization plugins
- [ ] Custom data source plugins
- [ ] Custom transform plugins
- [ ] Plugin marketplace/registry

### 6.2 Integrations

- [ ] dbt integration (model metadata)
- [ ] Airbyte/Fivetran (data pipelines)
- [ ] Jupyter notebook export
- [ ] Terraform provider

### 6.3 API & SDK

- [ ] Comprehensive REST API
- [ ] Go SDK
- [ ] TypeScript/JavaScript SDK
- [ ] Python SDK
- [ ] CLI for CI/CD workflows

### 6.4 Documentation

- [ ] User guide
- [ ] Admin guide
- [ ] API reference
- [ ] Plugin development guide
- [ ] Self-hosting guide

---

## Installation

### Static CLI (lightweight)

```bash
go install github.com/grokify/dashforge/cmd/dashforge@latest
```

### Full Server (includes database support)

```bash
go install github.com/grokify/dashforge/cmd/dashforge-server@latest
```

---

## Quick Start

### Static Dashboard (GitHub Pages)

```bash
# Create a dashboard
dashforge new my-dashboard

# Serve locally
dashforge serve ./my-dashboard

# Deploy: push JSON files to GitHub Pages
```

### Server Mode (with PostgreSQL)

```bash
# Start server with database
dashforge-server serve --db-url postgres://localhost/analytics

# Or use config file
dashforge-server serve --config config.yaml
```

---

## Priority Matrix

| Phase | Effort | Impact | Priority | Status |
|-------|--------|--------|----------|--------|
| 1.1 Schema & Validation | Low | High | P0 | ✅ Done |
| 1.3 Interactive Filters | Medium | High | P0 | ✅ Done |
| 1.7 CLI Tools | Low | Medium | P1 | Partial |
| 2.1 Server Foundation | Medium | High | P1 | Scaffolded |
| 2.2 PostgreSQL | Medium | High | P1 | Interface only |
| 1.2 Enhanced Viewer | Medium | High | P1 | Not started |
| 3.2 Query Builder UI | High | High | P2 | Not started |
| 4.1 Authentication | Medium | Medium | P2 | Scaffolded |
| 5.1 Scheduled Reports | Medium | Medium | P3 | Not started |
