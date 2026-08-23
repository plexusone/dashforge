# TRD — DashForge Technical Reference

**Initiative:** INIT-UIFORGE-001
**Status:** Draft
**Date:** 2026-07-23

## Architecture Overview

```text
PageSpec / ExperienceSpec (JSON IR)
        │
        ▼
   UISpec Validator
        │
        ▼
   Component Registry ◄── ComponentSpec manifests
        │
        ▼
   Layout Engine (responsive-grid, stack, split-pane, tabs)
        │
        ▼
   Data Binding Runtime
        │
        ▼
   Interaction Engine (event → condition → action)
        │
        ▼
   Design System (tokens, semantic variants)
        │
        ▼
   React Renderer
```

## Repository Structure

All work is consolidated in `plexusone/dashforge` as a monorepo. Consumer repos (`agentos`, `agentos-web`) import DashForge packages.

### dashforge/ (post-rename)

```text
dashforge/
├── cmd/
│   ├── dashforge/          # CLI (renamed from dashforge)
│   └── dashforge-server/   # HTTP server (renamed from dashforge-server)
│
├── uispec/               # NEW — canonical UISpec Go types (source of truth)
│   ├── page.go           # PageSpec: metadata, layout, components, interactions
│   ├── component.go      # ComponentInstance: type, version, properties, bindings
│   ├── layout.go         # LayoutSpec: responsive-grid, stack, split-pane, tabs
│   ├── interaction.go    # InteractionSpec: event → condition → action graph
│   ├── navigation.go     # NavigationSpec: routes, menus, breadcrumbs
│   ├── theme.go          # ThemeSpec: design tokens, profiles
│   ├── binding.go        # DataBindingSpec: source, operation, parameters
│   ├── capability.go     # CapabilitySpec: declared permissions per component
│   └── profile.go        # Profile constraints (dashboard, application, agent, etc.)
│
├── registry/             # NEW — component registry
│   ├── registry.go       # ComponentSpec manifest store (in-memory + persistent)
│   ├── manifest.go       # ComponentSpec: props schema, events, actions, constraints
│   ├── resolve.go        # Version resolution and lock-file logic
│   └── validate.go       # Validate a PageSpec against registered components
│
├── runtime/              # NEW — composition engine
│   ├── engine.go         # Load UISpec → resolve → produce render tree
│   ├── binding.go        # Evaluate data bindings and expressions
│   ├── interaction.go    # Route events through action graph
│   ├── visibility.go     # Evaluate visibility rules and permissions
│   └── state.go          # Runtime state management
│
├── dashboardir/          # EXISTING — Dashboard IR types (evolves into dashboard profile)
│   ├── dashboard.go      # Dashboard, Layout, Theme, Variable
│   ├── widget.go         # Widget, Position, DrillDown
│   ├── datasource.go     # DataSource, RefreshConfig
│   └── transform.go      # Transform pipeline
│
├── datasource/           # EXISTING — data source implementations
├── cube/                 # EXISTING — Cube.js integration
├── integration/          # EXISTING — third-party integrations
│
├── components/           # NEW — registered component implementations
│   ├── core/             # Buttons, cards, text, tabs, modals
│   ├── analytics/        # Charts, tables, metrics, filters (wraps dashboardir widgets)
│   ├── application/      # App shell, navigation, record views, forms
│   └── assistant/        # Assistant UI adapters (thread, composer, etc.)
│
├── schema/               # EXISTING + EXTENDED — generated JSON Schemas
│   ├── dashboard.schema.json    # Existing
│   ├── page.schema.json         # NEW — generated from uispec/page.go
│   ├── component.schema.json    # NEW — generated from registry/manifest.go
│   └── generate/                # Schema generation tooling
│
├── builder/              # EXISTING — Vite + React visual builder
├── viewer/               # EXISTING — embedded HTML viewer
├── ts/                   # EXISTING — TypeScript types
├── ent/                  # EXISTING — Ent ORM schemas
├── multiapp/             # EXISTING — multi-app support
│
├── examples/             # NEW — example PageSpecs
│   ├── dashboard/        # Dashboard profile examples
│   ├── agent-workspace/  # Agent profile examples
│   └── admin-console/    # Application profile examples
│
└── docs/
    ├── specs/            # PRD, TRD, PLAN, ROADMAP
    └── ...               # Existing docs
```

### agentos-web/ (consumer)

```text
agentos-web/
├── app/                  # Next.js app router
│   ├── chat/             # EXISTING — refactor to use DashForge page composition
│   ├── admin/            # EXISTING
│   ├── settings/         # EXISTING
│   └── ...
├── components/
│   ├── chat/             # EXISTING — 28 hand-coded components → migrate to DashForge
│   └── ...
├── specs/                # NEW — DashForge PageSpecs for AgentOS
│   ├── agent-workspace.page.json
│   ├── thread-view.page.json
│   ├── settings.page.json
│   └── ...
└── dashforge/              # NEW — AgentOS-specific DashForge component adapters
    ├── agentos.agent-selector.tsx
    ├── agentos.run-inspector.tsx
    └── agentos.model-selector.tsx
```

## UISpec Type System

### PageSpec (top-level)

```go
type PageSpec struct {
    APIVersion string            `json:"apiVersion"`           // "ui.plexusone.dev/v1"
    Kind       string            `json:"kind"`                 // "Page"
    Metadata   PageMetadata      `json:"metadata"`
    Profile    string            `json:"profile,omitempty"`    // "dashboard", "application", "agent"
    Context    map[string]string `json:"context,omitempty"`    // entity type, route params
    Layout     LayoutSpec        `json:"layout"`
    Components []ComponentInstance `json:"components"`
    Interactions []Interaction   `json:"interactions,omitempty"`
    Navigation *NavigationSpec   `json:"navigation,omitempty"`
    Theme      *ThemeRef         `json:"theme,omitempty"`
}
```

### ComponentInstance (placed on a page)

```go
type ComponentInstance struct {
    ID         string              `json:"id"`
    Type       string              `json:"type"`              // "analytics.line-chart", "assistant.thread"
    Version    string              `json:"version,omitempty"` // semver range
    Position   *Position           `json:"position,omitempty"`
    Properties map[string]any      `json:"properties,omitempty"`
    Data       map[string]Binding  `json:"data,omitempty"`
    Visibility *VisibilityRule     `json:"visibility,omitempty"`
    Slot       string              `json:"slot,omitempty"`    // for region/slot layouts
}
```

### ComponentSpec (manifest in registry)

```go
type ComponentSpec struct {
    ID                string              `json:"id"`            // "analytics.line-chart"
    Version           string              `json:"version"`       // "2.0.0"
    Category          string              `json:"category"`      // "visualization", "conversation", "form"
    Runtime           string              `json:"runtime"`       // "react"
    Entrypoint        string              `json:"entrypoint"`    // package path
    PropertiesSchema  json.RawMessage     `json:"propertiesSchema"`
    DataInputs        map[string]DataInput `json:"dataInputs,omitempty"`
    Events            map[string]EventDef  `json:"events,omitempty"`
    Actions           []string            `json:"actions,omitempty"`
    LayoutConstraints *LayoutConstraints   `json:"layoutConstraints,omitempty"`
    Capabilities      []string            `json:"capabilities,omitempty"`
    DesignSystem      *DesignSystemRef     `json:"designSystem,omitempty"`
}
```

## Layout Primitives

| Primitive | Description | Use case |
|---|---|---|
| `responsive-grid` | CSS Grid with breakpoints | Dashboards, record pages |
| `stack` | Vertical or horizontal flex | Forms, settings |
| `split-pane` | Resizable left/center/right | Agent workspace, IDE layouts |
| `tabs` | Tabbed content regions | Settings, detail views |
| `application-shell` | Navigation + header + main + inspector | Full applications |

## Component Namespaces

| Namespace | Scope | Examples |
|---|---|---|
| `core.*` | Platform primitives | `core.button`, `core.card`, `core.tabs`, `core.modal` |
| `analytics.*` | Dashboard/BI components | `analytics.line-chart`, `analytics.metric`, `analytics.table`, `analytics.filter` |
| `application.*` | Enterprise app components | `application.shell`, `application.nav`, `application.record-header` |
| `assistant.*` | Agent conversation (wraps Assistant UI) | `assistant.thread`, `assistant.composer`, `assistant.thread-list` |
| `agentos.*` | AgentOS domain components | `agentos.agent-selector`, `agentos.run-inspector`, `agentos.model-selector` |

## Assistant UI Integration

```text
DashForge ComponentSpec
        ↓
AssistantThread adapter component
        ↓
@assistant-ui/react primitives + runtime
        ↓
AgentOS backend APIs
```

Adapter components live in `dashforge/components/assistant/`. They:

1. Accept DashForge-standard properties and data bindings
2. Map them to Assistant UI's React primitives and runtime config
3. Connect to AgentOS via `ExternalStoreRuntime` (AgentOS owns conversation state)

The core DashForge packages (`uispec/`, `registry/`, `runtime/`) never import `@assistant-ui/react`. The dependency flows one way:

```text
components/assistant/ → @assistant-ui/react
components/assistant/ → uispec/
components/assistant/ → registry/
```

## Data Binding Model

Bindings connect components to data sources. Expressions use `${...}` syntax:

```json
{
  "data": {
    "customer": {
      "source": "customer-api",
      "operation": "getCustomer",
      "parameters": {
        "id": "${context.entityId}"
      }
    }
  }
}
```

Phase 1 supports static data and simple expression substitution. Phase 2 adds live data source connectors (reusing `datasource/` from DashForge).

## Interaction Model

Events flow through a declarative graph, not direct component coupling:

```json
{
  "interactions": [
    {
      "when": { "component": "revenue-chart", "event": "pointSelected" },
      "then": [
        { "action": "state.set", "parameters": { "path": "filters.month", "value": "${event.month}" } },
        { "action": "component.refresh", "target": "transaction-table" }
      ]
    }
  ]
}
```

## Design System Enforcement

Components expose semantic variants, not CSS. The design system decides colors, typography, spacing:

```json
{
  "type": "core.status-card",
  "properties": {
    "importance": "high",
    "density": "comfortable",
    "status": "warning"
  }
}
```

Compliance profiles control what's permitted:

| Profile | Components | Styling | Extensions |
|---|---|---|---|
| `strict` | Platform only | Tokens only | None |
| `managed` | Platform + approved extensions | Tokens + limited overrides | Approved packages |
| `open` | All + sandboxed custom | Custom within guardrails | Sandboxed |

## Module Rename Strategy

The Go module path changes from `github.com/plexusone/dashforge` to `github.com/plexusone/dashforge`. This is a breaking change for importers.

Steps:

1. Update `go.mod` module declaration
2. Rename `cmd/dashforge` → `cmd/dashforge`, `cmd/dashforge-server` → `cmd/dashforge-server`
3. Update all internal imports
4. Rename GitHub repo `plexusone/dashforge` → `plexusone/dashforge` (GitHub redirects old URLs)
5. Tag `v0.1.0` as the first DashForge release

Since the project is early and has few external consumers, the rename is low-risk.

## Testing Strategy

- **Unit tests** — UISpec types, registry validation, expression evaluation, layout resolution. Pure Go, no I/O.
- **Golden-file tests** — PageSpec JSON → rendered component tree snapshots. Catches unintended IR changes.
- **Integration tests** — Builder renders PageSpec in headless browser, validates component presence and layout.
- **Dashboard regression** — existing DashForge test suite runs under the dashboard profile unchanged.
- **agentos-web** — existing Playwright tests validate agent workspace renders from PageSpecs.

## Technology Stack

| Layer | Technology | Notes |
|---|---|---|
| Spec types | Go | Source of truth; `uispec/` package |
| Schema | JSON Schema (generated) | `invopop/jsonschema` + `schemago` lint |
| Server | Go + Chi | Existing DashForge server |
| ORM | Ent (MySQL/Dolt) | Existing DashForge schemas |
| Renderer | React 19 | DashForge React renderer package |
| Builder | Vite + React | Existing DashForge builder (evolve later) |
| Conversation | @assistant-ui/react ^0.7 | Used by agentos-web today |
| Frontend framework | Next.js 16 | agentos-web |
| Design system | TBD | Start with tokens in `uispec/theme.go`; extract later |
