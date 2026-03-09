# Dashforge

A JSON-first dashboard framework that starts simple with static hosting (GitHub Pages) and grows into a full Metabase-like analytics platform with an AI-powered visual builder.

## Features

- **Visual Dashboard Builder** - Drag-and-drop dashboard editor optimized for LLM agents
- **JSON Dashboard IR** - Non-polymorphic, AI-friendly dashboard definitions
- **Cube.js Semantic Layer** - Business-friendly queries with pre-built relationships
- **Static or Dynamic** - Start with static file hosting, graduate to PostgreSQL
- **Multi-Database Support** - Connect to PostgreSQL, MySQL, and more via plugin providers
- **Multi-tenant** - Row Level Security (RLS) for tenant isolation
- **SSO Authentication** - GitHub and Google OAuth support
- **ChartIR Integration** - Uses [echartify](https://github.com/grokify/echartify) for charts

## Quick Start

### Visual Dashboard Builder

The fastest way to create dashboards:

```bash
# Build the React dashboard builder
cd builder && npm install && npm run build && cd ..

# Start the server
go run ./cmd/dashforge-server serve --port 8080

# Open the builder
open http://localhost:8080/builder/
```

### Static Mode (No Server)

Open `viewer/index.html` in a browser with a dashboard URL:

```bash
cd dashforge
python3 -m http.server 8080
# Open http://localhost:8080/viewer/?dashboard=../examples/compliance-dashboard.json
```

### Server Mode

```bash
# Build the server
go build -o dashforge-server ./cmd/dashforge-server

# Run with PostgreSQL
export DATABASE_URL="postgres://user:pass@localhost:5432/dashforge?sslmode=disable"
export JWT_SECRET="your-secret-key"

./dashforge-server serve --port 8080 --auto-migrate
```

## Documentation

Full documentation is available at [docs/](docs/):

- [Getting Started](docs/getting-started.md)
- [Dashboard Builder](docs/builder.md) - Visual drag-and-drop editor
- [Dashboard IR Reference](docs/dashboard-ir.md)
- [Data Sources](docs/data-sources.md) - Database connections & providers
- [Cube.js Integration](docs/cube-integration.md) - Semantic data layer
- [AI Features](docs/ai-features.md) - LLM-powered dashboard generation
- [Server Configuration](docs/server-config.md)
- [Authentication](docs/authentication.md)
- [Multi-tenancy](docs/multi-tenancy.md)
- [API Reference](docs/api-reference.md)

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Dashforge                                │
├─────────────────────────────────────────────────────────────────┤
│  builder/               Visual dashboard builder (React)        │
│    ├── src/components/  Canvas, widgets, chart builder          │
│    ├── src/ai/          AI generation schemas & prompts         │
│    └── src/api/         Dashforge & Cube.js clients             │
├─────────────────────────────────────────────────────────────────┤
│  cube/                  Cube.js semantic layer                  │
│    └── model/cubes/     Data models (YAML)                      │
├─────────────────────────────────────────────────────────────────┤
│  cmd/dashforge/         Static CLI (validate, convert)          │
│  cmd/dashforge-server/  Full server with API                    │
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

Visualize compliance reports from [pipelineconductor](https://github.com/grokify/pipelineconductor):

```bash
# Generate compliance data
pipelineconductor check --users grokify --languages Go -o data/compliance.json

# View in dashboard
open viewer/index.html?dashboard=examples/compliance-dashboard.json

# Or use the visual builder
open http://localhost:8080/builder/
```

## License

MIT
