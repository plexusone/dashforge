# Dashforge

A JSON-first dashboard framework that starts simple with static hosting and grows into a full Metabase-like analytics platform with an AI-powered visual builder.

## Why Dashforge?

- **Visual Dashboard Builder** - Drag-and-drop editor optimized for both humans and LLM agents
- **JSON IR First** - Dashboards are defined in a non-polymorphic, AI-friendly JSON format that's easy to generate, validate, and version control
- **Cube.js Semantic Layer** - Business-friendly queries with pre-built relationships and AI context
- **Static to Dynamic** - Start with static file hosting (GitHub Pages), graduate to a full PostgreSQL-backed server when you need it
- **Multi-tenant Ready** - Built-in Row Level Security for SaaS deployments
- **SSO Authentication** - GitHub and Google OAuth out of the box
- **ChartIR Integration** - Uses [echartify](https://github.com/grokify/echartify) for declarative chart definitions

## Three Modes of Operation

### Builder Mode

Visual drag-and-drop dashboard creation:

- 12-column responsive grid layout
- Widget palette with charts, metrics, tables, and text
- Real-time chart preview with ECharts
- Cube.js query builder for semantic queries
- AI-powered widget generation from natural language

```bash
cd builder && npm install && npm run build && cd ..
./dashforge-server serve --port 8080
open http://localhost:8080/builder/
```

### Static Mode

Perfect for:

- Documentation sites
- GitHub Pages hosting
- Simple data visualization
- No backend required

```bash
# Just open in a browser
open viewer/index.html?dashboard=./examples/dashboard.json
```

### Server Mode

Full-featured analytics platform with:

- PostgreSQL data sources
- User authentication (OAuth)
- Multi-tenant isolation
- Dashboard CRUD API
- Saved queries

```bash
./dashforge-server serve --database-url postgres://... --auto-migrate
```

## Quick Links

- [Getting Started](getting-started.md) - Installation and first dashboard
- [Dashboard Builder](builder.md) - Visual editor guide
- [Cube.js Integration](cube-integration.md) - Semantic data layer
- [AI Features](ai-features.md) - LLM-powered generation
- [Dashboard IR](dashboard-ir.md) - JSON schema reference
- [Server Configuration](server-config.md) - Running the server
- [API Reference](api-reference.md) - REST API documentation

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Dashforge                                 │
├─────────────────────────────────────────────────────────────────┤
│  Builder Mode            │  Static Mode     │  Server Mode      │
│  ────────────            │  ───────────     │  ───────────      │
│  /builder/               │  viewer/         │  dashforge-server │
│  React + TypeScript      │  index.html      │  PostgreSQL + Ent │
│  Cube.js queries         │  JSON files      │  OAuth            │
│  AI generation           │  GitHub Pages    │  Multi-tenant RLS │
├─────────────────────────────────────────────────────────────────┤
│                    Dashboard IR (JSON)                           │
│                    ChartIR (echartify)                           │
│                    Cube.js Semantic Layer                        │
└─────────────────────────────────────────────────────────────────┘
```

## AI-First Design

Dashforge is designed from the ground up to work seamlessly with LLM agents:

- **Non-polymorphic JSON** - Dashboard IR uses flat structures without `oneOf`/`anyOf`, making it easy for AI to generate valid configurations
- **Semantic context** - Cube.js provides business-friendly names like "revenue" and "customer_count" instead of raw column names
- **Validated generation** - JSON Schema validation ensures AI-generated output is always valid
- **Dual rendering** - Same ChartIR spec works in the dashboard builder and chatbot responses

```
User: "Add a line chart showing revenue trends"
  ↓
AI generates ChartIR JSON
  ↓
Validation passes
  ↓
Widget added to dashboard
```
