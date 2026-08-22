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
- [ ] `RMI-UIFORGE-047` Documentation updates: analytics-catalog.md, README start command, examples
  - Depends on: `RMI-UIFORGE-045`
