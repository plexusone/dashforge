# ROADMAP — DashForge RMI Breakdown

**Initiative:** INIT-DASHFORGE-001
**Date:** 2026-07-23

RMI IDs use the repo slug where the work lands. Cross-repo items are attributed to the repo where the primary code change occurs.

---

## Phase 1: Foundation — UISpec & Module Rename

| ID | Title | Repo | Type | Required |
|---|---|---|---|---|
| RMI-DASHFORGE-001 | Rename module dashforge → dashforge | dashforge | refactor | yes |
| RMI-DASHFORGE-002 | Define UISpec core Go types (PageSpec, ComponentInstance, LayoutSpec) | dashforge | capability | yes |
| RMI-DASHFORGE-003 | Implement component registry with manifest validation | dashforge | capability | yes |
| RMI-DASHFORGE-004 | Define layout primitives (responsive-grid, stack, split-pane, tabs) | dashforge | capability | yes |
| RMI-DASHFORGE-005 | Define experience profiles (dashboard, application, agent, portal) | dashforge | capability | yes |
| RMI-DASHFORGE-006 | Generate JSON Schemas from UISpec Go types | dashforge | capability | yes |
| RMI-DASHFORGE-007 | Dashboard bridge: convert DashboardIR to UISpec PageSpec | dashforge | capability | yes |

### Dependencies

- RMI-DASHFORGE-002 → RMI-DASHFORGE-001 (types need new module path)
- RMI-DASHFORGE-003 → RMI-DASHFORGE-002 (registry validates against UISpec types)
- RMI-DASHFORGE-004 → RMI-DASHFORGE-002 (layout types are part of UISpec)
- RMI-DASHFORGE-005 → RMI-DASHFORGE-002 (profiles reference UISpec types)
- RMI-DASHFORGE-006 → RMI-DASHFORGE-002 (schemas generated from types)
- RMI-DASHFORGE-007 → RMI-DASHFORGE-002, RMI-DASHFORGE-004 (bridge produces PageSpecs with layouts)

---

## Phase 2: Component Library & Dashboard Parity

| ID | Title | Repo | Type | Required |
|---|---|---|---|---|
| RMI-DASHFORGE-008 | Register core components (card, text, tabs, button, modal) | dashforge | capability | yes |
| RMI-DASHFORGE-009 | Register analytics components (line-chart, bar-chart, metric, table, filter, gauge) | dashforge | capability | yes |
| RMI-DASHFORGE-010 | Implement UISpec React renderer | dashforge | capability | yes |
| RMI-DASHFORGE-011 | Dashboard profile enforcement and validation | dashforge | quality | yes |
| RMI-DASHFORGE-012 | Convert DashForge example dashboards to UISpec PageSpecs | dashforge | quality | no |
| RMI-DASHFORGE-013 | Golden-file tests for PageSpec → component tree | dashforge | quality | yes |

### Dependencies

- RMI-DASHFORGE-008 → RMI-DASHFORGE-003 (components need registry)
- RMI-DASHFORGE-009 → RMI-DASHFORGE-003, RMI-DASHFORGE-007 (analytics wraps dashboard widgets)
- RMI-DASHFORGE-010 → RMI-DASHFORGE-003, RMI-DASHFORGE-008 (renderer needs registry + components)
- RMI-DASHFORGE-011 → RMI-DASHFORGE-005, RMI-DASHFORGE-009 (enforcement needs profiles + components)
- RMI-DASHFORGE-012 → RMI-DASHFORGE-010 (examples need renderer)
- RMI-DASHFORGE-013 → RMI-DASHFORGE-010 (golden tests need renderer)

---

## Phase 3: Assistant UI Integration

| ID | Title | Repo | Type | Required |
|---|---|---|---|---|
| RMI-DASHFORGE-014 | Implement assistant.thread component adapter | dashforge | capability | yes |
| RMI-DASHFORGE-015 | Implement assistant.composer and assistant.thread-list adapters | dashforge | capability | yes |
| RMI-DASHFORGE-016 | Implement assistant.tool-call and assistant.run-status adapters | dashforge | capability | no |
| RMI-DASHFORGE-017 | ExternalStoreRuntime adapter for AgentOS backend | dashforge | capability | yes |
| RMI-DASHFORGE-018 | Define agent experience profile | dashforge | capability | yes |
| RMI-DASHFORGE-019 | Sample agent workspace PageSpec with functional streaming | dashforge | quality | yes |

### Dependencies

- RMI-DASHFORGE-014 → RMI-DASHFORGE-010 (adapters need renderer)
- RMI-DASHFORGE-015 → RMI-DASHFORGE-014 (build on thread adapter patterns)
- RMI-DASHFORGE-016 → RMI-DASHFORGE-014 (same adapter pattern)
- RMI-DASHFORGE-017 → RMI-DASHFORGE-014 (runtime connects to thread)
- RMI-DASHFORGE-018 → RMI-DASHFORGE-005 (extends profile system)
- RMI-DASHFORGE-019 → RMI-DASHFORGE-014, RMI-DASHFORGE-015, RMI-DASHFORGE-017 (sample needs all assistant components)

---

## Phase 4: AgentOS Consumption

| ID | Title | Repo | Type | Required |
|---|---|---|---|---|
| RMI-AGENTOS-001 | Register agentos.agent-selector component | agentos-web | capability | yes |
| RMI-AGENTOS-002 | Register agentos.run-inspector and agentos.model-selector components | agentos-web | capability | no |
| RMI-AGENTOS-003 | Agent workspace PageSpec (navigation + thread list + thread + inspector) | agentos-web | capability | yes |
| RMI-AGENTOS-004 | Thread view PageSpec | agentos-web | capability | yes |
| RMI-AGENTOS-005 | Settings page PageSpec | agentos-web | capability | no |
| RMI-AGENTOS-006 | Next.js page routing integration with DashForge renderer | agentos-web | capability | yes |
| RMI-AGENTOS-007 | Migrate chat components from hand-coded to DashForge composition | agentos-web | refactor | yes |
| RMI-AGENTOS-008 | AgentOS API data binding adapters | agentos | capability | yes |

### Dependencies

- RMI-AGENTOS-001 → RMI-DASHFORGE-010 (needs DashForge renderer)
- RMI-AGENTOS-002 → RMI-AGENTOS-001 (same pattern)
- RMI-AGENTOS-003 → RMI-DASHFORGE-019, RMI-AGENTOS-001 (needs assistant + agentos components)
- RMI-AGENTOS-004 → RMI-DASHFORGE-014, RMI-AGENTOS-006 (needs thread adapter + routing)
- RMI-AGENTOS-005 → RMI-DASHFORGE-008, RMI-AGENTOS-006 (needs core components + routing)
- RMI-AGENTOS-006 → RMI-DASHFORGE-010 (routing loads renderer)
- RMI-AGENTOS-007 → RMI-AGENTOS-003, RMI-AGENTOS-004 (migrate after new pages work)
- RMI-AGENTOS-008 → RMI-DASHFORGE-017 (bindings extend runtime adapter)

---

## Phase 5: Interaction Engine & Data Binding

| ID | Title | Repo | Type | Required |
|---|---|---|---|---|
| RMI-DASHFORGE-020 | Expression evaluator (${...} syntax) in Go and TypeScript | dashforge | capability | yes |
| RMI-DASHFORGE-021 | Declarative interaction engine (event → condition → action graph) | dashforge | capability | yes |
| RMI-DASHFORGE-022 | Page-level shared state management | dashforge | capability | yes |
| RMI-DASHFORGE-023 | Data source connectors via UISpec bindings | dashforge | capability | yes |
| RMI-DASHFORGE-024 | Dashboard drill-down interaction (chart click → table filter) | dashforge | quality | yes |

### Dependencies

- RMI-DASHFORGE-020 → RMI-DASHFORGE-002 (expressions reference UISpec binding syntax)
- RMI-DASHFORGE-021 → RMI-DASHFORGE-020 (interactions evaluate expressions)
- RMI-DASHFORGE-022 → RMI-DASHFORGE-021 (state is mutated by interaction actions)
- RMI-DASHFORGE-023 → RMI-DASHFORGE-020 (bindings use expression evaluator)
- RMI-DASHFORGE-024 → RMI-DASHFORGE-021, RMI-DASHFORGE-023, RMI-DASHFORGE-009 (needs interactions + data + analytics components)

---

## Phase 6: Builder Evolution & AI Generation

| ID | Title | Repo | Type | Required |
|---|---|---|---|---|
| RMI-DASHFORGE-025 | Upgrade builder to load/save UISpec PageSpecs | dashforge | capability | yes |
| RMI-DASHFORGE-026 | Component palette from registry | dashforge | capability | yes |
| RMI-DASHFORGE-027 | Property editor from ComponentSpec manifests | dashforge | capability | yes |
| RMI-DASHFORGE-028 | AI PageSpec generation with validation feedback loop | dashforge | capability | no |
| RMI-DASHFORGE-029 | PageSpec diffing for code review | dashforge | quality | no |

### Dependencies

- RMI-DASHFORGE-025 → RMI-DASHFORGE-010 (builder needs renderer)
- RMI-DASHFORGE-026 → RMI-DASHFORGE-025, RMI-DASHFORGE-003 (palette reads registry)
- RMI-DASHFORGE-027 → RMI-DASHFORGE-026 (property editor extends palette selection)
- RMI-DASHFORGE-028 → RMI-DASHFORGE-006 (AI generates against schema)
- RMI-DASHFORGE-029 → RMI-DASHFORGE-002 (diffing operates on UISpec types)

---

## Summary

| Phase | RMIs | Required | Repos |
|---|---|---|---|
| 1 — Foundation | 7 | 7 | dashforge |
| 2 — Components & Dashboard | 6 | 5 | dashforge |
| 3 — Assistant UI | 6 | 5 | dashforge |
| 4 — AgentOS Consumption | 8 | 6 | agentos-web (6), agentos (2) |
| 5 — Interactions & Data | 5 | 5 | dashforge |
| 6 — Builder & AI | 5 | 3 | dashforge |
| **Total** | **37** | **31** | |
