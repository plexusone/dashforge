# UIForge

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/uiforge/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/uiforge/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/uiforge/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/uiforge/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/uiforge/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/uiforge/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/uiforge
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/uiforge
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/uiforge
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fuiforge
 [loc-svg]: https://tokei.rs/b1/github/plexusone/uiforge
 [repo-url]: https://github.com/plexusone/uiforge
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/uiforge/blob/main/LICENSE

A JSON-first dashboard framework that starts simple with static hosting (GitHub Pages) and grows into a full Metabase-like analytics platform with an AI-powered visual builder.

## Features

### Core Platform

- 🎨 **Visual Dashboard Builder** - Drag-and-drop dashboard editor optimized for LLM agents
- 📄 **JSON Dashboard IR** - Non-polymorphic, AI-friendly dashboard definitions
- 🔗 **Cube.js Semantic Layer** - Business-friendly queries with pre-built relationships
- ⚡ **Static or Dynamic** - Start with static file hosting, graduate to PostgreSQL
- 🗄️ **Multi-Database Support** - Connect to PostgreSQL, MySQL, and more via plugin providers
- ❓ **Saved Questions** - Metabase-style read-only GrokifyQL questions with formatting, syntax highlighting, field value browsing, CSV/XLSX export, and dashboard reuse
- 🤖 **Agent-Ready Analytics** - Integrates with OmniAgent marketplace primitives while UIForge owns query, field-value, and dashboard capabilities
- 🔑 **Secret References** - Analytics source credentials should use OmniVault/VaultGuard references instead of raw DSNs in catalog storage, with OmniVault SQL store as the encrypted-in-DB fallback
- 🏢 **Multi-tenant** - Row Level Security (RLS) for tenant isolation
- 🔐 **SSO Authentication** - GitHub and Google OAuth support
- 📊 **ChartIR Integration** - Uses [echartify](https://github.com/grokify/echartify) for charts

### Template Marketplace

- 📦 **Dashboard Templates** - Publish and sell reusable dashboard templates
- 🏪 **Publisher System** - Organizations can become publishers with creator roles
- ✅ **Template Licensing** - Seat-based licensing for purchased templates
- 🔄 **Version Control** - Template versioning with auto-update options
- 🖼️ **Preview & Screenshots** - Gallery views for template discovery

### Authorization

- 🔀 **Dual-Mode Auth** - Simple role hierarchy or SpiceDB for fine-grained control
- 🧭 **Analytics Policy** - SystemForge/SpiceDB-backed GrokifyQL policy for dataset and field permissions
- 👤 **Publisher Roles** - Owner, Admin, Creator, Reviewer hierarchies
- 👥 **Consumer Roles** - Owner, Admin, Editor, Viewer hierarchies
- 🎯 **Resource Permissions** - Granular control over dashboards, queries, alerts, integrations

## Quick Start

### Visual Dashboard Builder

The fastest way to create dashboards:

```bash
# Build the React dashboard builder
cd builder && npm install && npm run build && cd ..

# Start the server
go run ./cmd/uiforge-server serve --port 8080

# Open the builder
open http://localhost:8080/builder/
```

### Static Mode (No Server)

Open `viewer/index.html` in a browser with a dashboard URL:

```bash
cd uiforge
python3 -m http.server 8080
# Open http://localhost:8080/viewer/?dashboard=../examples/compliance-dashboard.json
```

The viewer has a light/dark theme toggle (top-right, persisted via `localStorage`) and caps numeric display in tooltips, axis labels, and metric tiles to 2 decimal places.

### Server Mode

```bash
# Build the server
go build -o uiforge-server ./cmd/uiforge-server

# Run with Dolt (local default — connects to a dolt sql-server, creating the
# database if needed; matches the OmniRoadmap/VisionStudio local pattern)
./uiforge-server serve --port 8080 --auto-migrate \
  --db-url 'mysql://root@127.0.0.1:13307/uiforge'

# Or with PostgreSQL (required for Row Level Security / --enable-rls)
export DATABASE_URL="postgres://user:pass@localhost:5432/uiforge?sslmode=disable"
export JWT_SECRET="your-secret-key"

./uiforge-server serve --port 8080 --auto-migrate
```

### Analytics Sources

UIForge connects to multiple analytics sources at once, like Metabase. Sources
are added at runtime — via the **Data sources** panel in the Question builder
(`/builder/?mode=questions`) or the API — and persist across restarts. Only an
OmniVault secret reference is stored, never a raw DSN:

```bash
export UIFORGE_OMNIROADMAP_DSN='root:@tcp(127.0.0.1:13307)/omniroadmap'
./uiforge-server serve --address 127.0.0.1:13319

# One-time: register the source (or use the builder UI)
curl -X POST http://127.0.0.1:13319/api/v1/analytics/sources \
  -H 'Content-Type: application/json' \
  -d '{"id":"omniroadmap","name":"OmniRoadmap","connector":"omniroadmap",
       "dsnRef":"env://UIFORGE_OMNIROADMAP_DSN","enabled":true}'
```

See [Analytics Catalog](docs/analytics-catalog.md) for the catalog model,
management API, and credential policy.

## Documentation

Full documentation is available at [docs/](docs/):

- [Getting Started](docs/getting-started.md)
- [Dashboard Builder](docs/builder.md) - Visual drag-and-drop editor
- [Dashboard IR Reference](docs/dashboard-ir.md)
- [Data Sources](docs/data-sources.md) - Database connections & providers
- [Analytics Catalog](docs/analytics-catalog.md) - Queryable sources, datasets, fields, and Questions
- [Analytics Authorization](docs/analytics-authorization.md) - GrokifyQL policy with SystemForge and SpiceDB
- [Agent Integration](docs/agent-integration.md) - OmniAgent marketplace boundary and UIForge analytics capabilities
- [Cube.js Integration](docs/cube-integration.md) - Semantic data layer
- [AI Features](docs/ai-features.md) - LLM-powered dashboard generation
- [Server Configuration](docs/server-config.md)
- [Authentication](docs/authentication.md)
- [Multi-tenancy](docs/multi-tenancy.md)
- [API Reference](docs/api-reference.md)

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        UIForge                                  │
├─────────────────────────────────────────────────────────────────┤
│  builder/               Visual dashboard builder (React)        │
│    ├── src/components/  Canvas, widgets, chart builder          │
│    ├── src/ai/          AI generation schemas & prompts         │
│    └── src/api/         UIForge & Cube.js clients               │
├─────────────────────────────────────────────────────────────────┤
│  cube/                  Cube.js semantic layer                  │
│    └── model/cubes/     Data models (YAML)                      │
├─────────────────────────────────────────────────────────────────┤
│  cmd/uiforge/         Static CLI (validate, convert)            │
│  cmd/uiforge-server/  Full server with API                      │
├─────────────────────────────────────────────────────────────────┤
│  dashboardir/           Dashboard JSON schema & types           │
│  viewer/                Embedded static HTML/JS viewer          │
├─────────────────────────────────────────────────────────────────┤
│  datasource/            Plugin-style data source providers      │
│    ├── providers/       PostgreSQL, MySQL implementations       │
│    ├── manager.go       Connection pool management              │
│    └── query.go         Query execution engine                  │
├─────────────────────────────────────────────────────────────────┤
│  internal/server/                                               │
│    ├── api/             REST API handlers                       │
│    ├── analytics/       Queryable source catalog and providers  │
│    ├── auth/            JWT + OAuth (GitHub, Google)            │
│    ├── db/              PostgreSQL with Ent ORM                 │
│    └── middleware/      Tenant context, logging                 │
├─────────────────────────────────────────────────────────────────┤
│  ent/                   Ent schema & generated code             │
│    └── schema/          User, Dashboard, Tenant, etc.           │
└─────────────────────────────────────────────────────────────────┘
```

## Dashboard Builder

The visual builder provides a Metabase-style drag-and-drop interface:

- **Widget Palette** - Drag charts, metrics, tables, and text onto the canvas
- **12-Column Grid** - Responsive layout with snap-to-grid positioning
- **Chart Builder** - Visual configuration for line, bar, pie, scatter, and area charts
- **Query Builder** - Connect to Cube.js for semantic queries
- **AI Integration** - Generate widgets and dashboards from natural language
- **Question Builder** - Create saved GrokifyQL Questions, browse field values,
  run results, export CSV/XLSX, and reuse Questions as dashboard widgets

```
┌─────────────────────────────────────────────────────────────────┐
│                   Dashboard Builder UI                          │
│         (React + TypeScript + react-grid-layout)                │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ Widget      │  │ Canvas      │  │ Properties Panel        │  │
│  │ Palette     │  │ (Grid)      │  │ ├── Chart Builder       │  │
│  │ ├── Chart   │  │             │  │ ├── Query Builder       │  │
│  │ ├── Metric  │  │  [Widget]   │  │ └── Style Editor        │  │
│  │ ├── Table   │  │  [Widget]   │  │                         │  │
│  │ └── Text    │  │  [Widget]   │  │                         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Dashboard Example

```json
{
  "id": "sales-dashboard",
  "title": "Sales Overview",
  "layout": { "type": "grid", "columns": 12, "rowHeight": 80 },
  "dataSources": [
    {
      "id": "sales",
      "type": "postgres",
      "query": "SELECT date, SUM(amount) as total FROM sales GROUP BY date"
    }
  ],
  "widgets": [
    {
      "id": "revenue-chart",
      "type": "chart",
      "position": { "x": 0, "y": 0, "w": 8, "h": 4 },
      "dataSourceId": "sales",
      "config": {
        "geometry": "line",
        "encodings": { "x": "date", "y": "total" },
        "style": { "smooth": true, "showLegend": true }
      }
    }
  ]
}
```

## Development

```bash
# Build all binaries
go build ./...

# Build the dashboard builder
cd builder && npm install && npm run build && cd ..

# Run tests
go test -v ./...

# Lint
golangci-lint run

# Generate Ent code (after schema changes)
go generate ./ent

# Start Cube.js (optional, for semantic queries)
cd cube && npm install && npm run dev
```

## Integration with PipelineConductor

Visualize compliance reports from [pipelineconductor](https://github.com/plexusone/pipelineconductor):

```bash
# Generate compliance data
pipelineconductor check --users plexusone --languages Go -o data/compliance.json

# View in dashboard
open viewer/index.html?dashboard=examples/compliance-dashboard.json

# Or use the visual builder
open http://localhost:8080/builder/
```

## License

MIT
