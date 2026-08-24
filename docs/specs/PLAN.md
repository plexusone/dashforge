# PLAN — DashForge Build Order

**Initiative:** INIT-UIFORGE-001
**Status:** Draft
**Date:** 2026-07-23

## Phase Overview

| Phase | Theme | Repos | Exit Criteria |
|---|---|---|---|
| 1 | Foundation — UISpec & Module Rename | dashforge | Go module renamed, UISpec types defined, component registry operational, dashboard profile passes existing tests |
| 2 | Component Library & Dashboard Parity | dashforge | Core + analytics components registered with manifests, existing DashForge dashboards render via UISpec PageSpecs |
| 3 | Assistant UI Integration | dashforge | `assistant.*` components registered, sample agent workspace PageSpec renders thread + composer + thread-list |
| 4 | AgentOS Consumption | agentos, agentos-web | Agent workspace, thread view, and settings pages composed from DashForge PageSpecs; zero hand-coded page layout in agentos-web |
| 5 | Interaction Engine & Data Binding | dashforge | Cross-component events, expression evaluation, and live data binding working for dashboard drill-down and agent workspace interactions |
| 6 | Builder Evolution & AI Generation | dashforge | Visual builder supports UISpec PageSpecs (not just DashboardIR); LLM produces valid PageSpecs that render correctly |

---

## Phase 1: Foundation — UISpec & Module Rename

**Goal:** Establish the UISpec type system and component registry as the new foundation, while preserving DashForge compatibility.

**Work:**

1. **Module rename** — `dashforge` → `uiforge` in `go.mod`, all internal imports, CLI commands, and GitHub repo name.
2. **UISpec Go types** — `uispec/` package with `PageSpec`, `ComponentInstance`, `LayoutSpec`, `ThemeRef`, `Binding`, `Interaction`, `VisibilityRule`, `Profile`.
3. **Component registry** — `registry/` package with `ComponentSpec` manifests, in-memory store, validation (verify a PageSpec references only registered components with valid properties).
4. **Layout primitives** — `responsive-grid`, `stack`, `split-pane`, `tabs` layout types in `uispec/layout.go`.
5. **Profile definitions** — dashboard, application, agent profile constraint sets.
6. **Schema generation** — generate `page.schema.json` and `component.schema.json` from Go types.
7. **Dashboard bridge** — adapter that converts existing `dashboardir.Dashboard` to `uispec.PageSpec` with dashboard profile, ensuring backward compatibility.

**Exit criteria:**

- `go build ./...` passes with new module path
- Existing tests pass (dashboard functionality preserved)
- `prismctl registry` shows `github.com/plexusone/dashforge`
- A hand-written PageSpec (JSON) validates against the registry and resolves all component references

**Risks:**

- Import path churn for existing consumers → mitigated by early rename while consumer count is low
- UISpec over-design before validation → keep types minimal; add fields when real components need them

---

## Phase 2: Component Library & Dashboard Parity

**Goal:** Register concrete component implementations with full manifests. Existing DashForge dashboards render through the new UISpec pipeline.

**Work:**

1. **Core components** — register `core.card`, `core.text`, `core.tabs`, `core.button`, `core.modal` with ComponentSpec manifests.
2. **Analytics components** — register `analytics.line-chart`, `analytics.bar-chart`, `analytics.metric`, `analytics.table`, `analytics.filter`, `analytics.gauge` wrapping existing DashForge widget types.
3. **Dashboard profile enforcement** — validate that dashboard-profile PageSpecs only use permitted component namespaces and layout types.
4. **React renderer** — minimal UISpec → React tree renderer that reads a PageSpec and renders registered components. Start in `ts/` or `builder/`.
5. **Golden-file tests** — PageSpec JSON → expected component tree snapshots for each registered component.
6. **Example PageSpecs** — convert 2-3 existing DashForge example dashboards to UISpec PageSpecs in `examples/dashboard/`.

**Exit criteria:**

- 10+ core components and 5+ analytics components registered
- Existing DashForge example dashboards render identically via UISpec pipeline
- `dashboard.schema.json` and `page.schema.json` both validate their respective example files
- React renderer produces correct output for all registered components

**Risks:**

- Widget config diversity (ECharts IR, TableConfig, MetricConfig) complicates ComponentSpec property schemas → use `json.RawMessage` for type-specific config initially, add typed schemas incrementally
- Renderer scope creep → Phase 2 renderer handles static rendering only; no interactions or live data

---

## Phase 3: Assistant UI Integration

**Goal:** Wrap Assistant UI primitives as DashForge components so agent conversation UIs can be composed declaratively.

**Work:**

1. **Assistant component adapters** — implement `assistant.thread`, `assistant.composer`, `assistant.thread-list`, `assistant.tool-call`, `assistant.run-status` in `components/assistant/`.
2. **ComponentSpec manifests** — each assistant component gets a full manifest (props schema, events, actions, layout constraints).
3. **Runtime integration** — `ExternalStoreRuntime` adapter connecting Assistant UI to AgentOS backend APIs (messages, streaming, tool calls).
4. **Agent profile** — define permitted components, layouts (`split-pane`, `application-shell`), and conventions for agent workspaces.
5. **Sample agent workspace PageSpec** — a working PageSpec that composes `assistant.thread`, `assistant.thread-list`, and `assistant.composer` in a split-pane layout.

**Exit criteria:**

- 5+ assistant components registered with manifests
- Sample agent workspace PageSpec renders a functional conversation UI
- Assistant UI streaming works through DashForge composition (messages appear in real time)
- Core DashForge packages have zero imports from `@assistant-ui/react`

**Risks:**

- Assistant UI version coupling → pin to `^0.7`, adapter handles version differences
- Streaming complexity → ExternalStoreRuntime is designed for this; follow Assistant UI's documented pattern
- State management boundary unclear → AgentOS owns all conversation state; DashForge only renders

---

## Phase 4: AgentOS Consumption

**Goal:** Refactor `agentos-web` to compose pages from DashForge PageSpecs. AgentOS becomes the reference implementation.

**Work:**

1. **AgentOS domain components** — register `agentos.agent-selector`, `agentos.run-inspector`, `agentos.model-selector`, `agentos.context-viewer` as DashForge components.
2. **Agent workspace PageSpec** — full workspace: navigation + thread list + thread + inspector panels.
3. **Thread view PageSpec** — conversation thread with composer, tool-call display, and attachment support.
4. **Settings PageSpec** — settings page with tabs, forms, and model configuration.
5. **Page routing** — Next.js app router loads PageSpecs and passes them to DashForge React renderer.
6. **AgentOS API bindings** — data binding adapters for AgentOS's conversation, agent, and model APIs.
7. **Migration** — incrementally replace hand-coded chat components (28 files in `components/chat/`) with DashForge composition.

**Exit criteria:**

- Agent workspace, thread view, and settings pages render from PageSpecs
- All page layout is defined in PageSpecs; React code exists only in component implementations
- Existing agentos-web functionality preserved (collaborative chat, model selector, reactions, etc.)
- Playwright tests pass against the new composition

**Risks:**

- Feature parity during migration → migrate one page at a time; keep old components until the new version passes all tests
- Performance regression → DashForge adds an interpretation layer; profile and optimize the render path
- 28 chat components have deep interdependencies → map dependencies before starting migration

---

## Phase 5: Interaction Engine & Data Binding

**Goal:** Cross-component interactions and live data binding make dashboards and workspaces interactive without custom code.

**Work:**

1. **Expression evaluator** — `${context.entityId}`, `${event.month}`, `${data.customer.name}` syntax evaluation in Go and TypeScript.
2. **Event routing** — declarative interaction rules (`when` → `then` action graph) evaluated at runtime.
3. **State management** — shared page-level state that components can read from and write to.
4. **Data source connectors** — reuse DashForge's `datasource/` package (SQL, REST, static) through UISpec bindings.
5. **Live refresh** — data source polling and WebSocket push for real-time dashboards.
6. **Dashboard drill-down** — clicking a chart data point filters a table (classic dashboard interaction, expressed declaratively).

**Exit criteria:**

- Expression evaluation handles all supported syntax in Go and TypeScript
- Dashboard drill-down works: chart click → filter → table refresh, all defined in PageSpec interactions
- Agent workspace interactions work: thread selection → message list update
- No imperative event wiring in PageSpecs; all interactions are declarative

**Risks:**

- Expression language creep → keep it to property access and simple conditions; no Turing-completeness
- State synchronization between Go runtime and React → single source of truth in the browser; Go validates at compile/deploy time

---

## Phase 6: Builder Evolution & AI Generation

**Goal:** The visual builder understands UISpec (not just DashboardIR) and AI agents can generate valid PageSpecs.

**Work:**

1. **Builder upgrade** — extend the existing Vite builder to load/save UISpec PageSpecs (not just DashboardIR).
2. **Component palette** — builder shows available components from the registry, grouped by namespace.
3. **Property editor** — generate property editors from ComponentSpec manifests (JSON Schema → form).
4. **AI generation** — prompt templates and examples that let LLMs produce valid PageSpecs. Validation feedback loop: generate → validate → fix → render.
5. **PageSpec diffing** — human-readable diffs of PageSpec changes for code review.

**Exit criteria:**

- Builder can load, edit, and save a UISpec PageSpec with components from any namespace
- An LLM (Claude) can generate a valid dashboard PageSpec from a natural-language description in <3 attempts
- Generated PageSpecs pass validation and render correctly
- PageSpec changes produce meaningful diffs in `git diff`

**Risks:**

- Builder scope explosion → Phase 6 builder supports layout + component placement + property editing only; interaction editor is post-Phase 6
- AI hallucination of non-existent components → validation feedback loop catches this; registry is the authority

---

## Dependency Order

```text
Phase 1 ──▶ Phase 2 ──▶ Phase 3 ──▶ Phase 4
                │                       │
                ▼                       ▼
             Phase 5 ◄──────────────────┘
                │
                ▼
             Phase 6
```

Phase 1 is prerequisite for everything. Phases 2 and 3 can partially overlap (different component namespaces). Phase 4 depends on Phase 3 (assistant components must exist). Phase 5 can start after Phase 2 (dashboards need interactions) and benefits from Phase 4 learnings. Phase 6 depends on Phase 5 (builder needs interactions).
