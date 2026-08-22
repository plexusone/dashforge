# ROADMAP — Persistent Analytics Sources — Metabase-Style Multi-Source Catalog with OmniVault

**Initiative:** `INIT-UIFORGE-002`
**Repository:** `github.com/plexusone/uiforge`

## Phase 1 — Source Model, Secret Resolution, Persistence

**Theme:** Backend foundation: config model, OmniVault resolver, dual-backend store

- [ ] `RMI-UIFORGE-039` SourceConfig type, validation, and connector registry with omniroadmap factory
- [ ] `RMI-UIFORGE-040` OmniVault v0.5.0 resolver integration with env:// and file:// providers
- [ ] `RMI-UIFORGE-041` SourceStore interface and JSON file store fallback
  - Depends on: `RMI-UIFORGE-039`
- [ ] `RMI-UIFORGE-042` Ent AnalyticsSource schema, codegen, and ent-backed store
  - Depends on: `RMI-UIFORGE-041`

## Phase 2 — Dynamic Service, API, Flag Removal

**Theme:** Runtime add/remove of sources; server starts with no per-source flags

- [ ] `RMI-UIFORGE-043` Dynamic analytics.Service with LoadAll, Add/Update/Remove/TestSource, and catalog ID re-tagging
  - Depends on: `RMI-UIFORGE-039`
  - Depends on: `RMI-UIFORGE-040`
  - Depends on: `RMI-UIFORGE-041`
- [ ] `RMI-UIFORGE-044` Analytics sources CRUD, test, and connectors API endpoints
  - Depends on: `RMI-UIFORGE-043`
- [ ] `RMI-UIFORGE-045` Server wiring: startup LoadAll and removal of --omniroadmap-dsn flag
  - Depends on: `RMI-UIFORGE-044`

## Phase 3 — Builder UI and Documentation

**Theme:** Metabase-style source management in the Question workspace

- [ ] `RMI-UIFORGE-046` AnalyticsSourcePanel in Question workspace with add/edit/test/remove and catalog refresh
  - Depends on: `RMI-UIFORGE-044`
- [x] `RMI-UIFORGE-049` Shared AppNav: consistent Dashboards/Questions/Data sources navigation across builder modes
- [ ] `RMI-UIFORGE-047` Documentation updates: analytics-catalog.md, README start command, examples
  - Depends on: `RMI-UIFORGE-045`

## Phase 4 — Dolt Metadata DB and Shared Dolt Library

**Theme:** Dolt as UIForge's local metadata DB; consolidate shared Dolt wiring into grokify/godolt

Note: `RMI-GODOLT-001`, `RMI-OMNIROADMAP-010`, and `RMI-VISIONSTUDIO-549` are
cross-repo RMIs created directly in VisionStudio (roadmap import assigns a
single repository, so they are listed here for reference only and are not
managed by `roadmap import` of this file).

- [ ] `RMI-GODOLT-001` (repo `github.com/grokify/godolt`) Server lifecycle package: EnsureServer, InitDatabase, DSN helpers, Commit/Status client methods
- [ ] `RMI-UIFORGE-048` Dolt metadata DB: implement mysql:// in internal/server/db via Ent MySQL dialect and godolt
  - Depends on: `RMI-GODOLT-001`
- [ ] `RMI-OMNIROADMAP-010` (repo `github.com/grokify/omniroadmap`) Refactor store/doltstore.go to consume godolt server lifecycle helpers
  - Depends on: `RMI-GODOLT-001`
- [ ] `RMI-VISIONSTUDIO-549` (repo `github.com/ProductBuildersHQ/visionstudio`) Refactor Dolt server wiring to consume godolt lifecycle helpers
  - Depends on: `RMI-GODOLT-001`

## Phase 5 — Navigation Shell

**Theme:** Consolidate builder chrome into a collapsible left rail with contextual section columns; top bars carry document tools only

- [ ] `RMI-UIFORGE-050` AppShell: collapsible left rail with product sections wrapping both builder modes
- [ ] `RMI-UIFORGE-051` Dashboards section: persistent list column replacing the gallery modal
  - Depends on: `RMI-UIFORGE-050`
- [ ] `RMI-UIFORGE-052` Unified Data section: analytics sources and SQL connections as tabs of one view
  - Depends on: `RMI-UIFORGE-050`
- [ ] `RMI-UIFORGE-053` Top bar cleanup: document chrome only, Variables moved into dashboard settings
  - Depends on: `RMI-UIFORGE-050`
