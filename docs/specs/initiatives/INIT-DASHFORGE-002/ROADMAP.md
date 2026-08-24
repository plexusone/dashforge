# ROADMAP — Connector Extensibility — Protocol Drivers and the Analytics Provider Contract

**Initiative:** `INIT-DASHFORGE-002`
**Repository:** `github.com/plexusone/dashforge`

`RMI-OMNIROADMAP-023` is a cross-repo RMI created directly in VisionStudio
(roadmap import assigns a single repository; listed for reference only).

## Phase 1 — Tier 1: Generic SQL Connector

**Theme:** Any Postgres/MySQL/Dolt database becomes a source via configuration alone

- [ ] `RMI-DASHFORGE-008` Generic sql connector: driver selection and INFORMATION_SCHEMA catalog introspection
- [ ] `RMI-DASHFORGE-009` SQL dialect query path: read-only allowlist, LIMIT clamp, sanitized errors
  - Depends on: `RMI-DASHFORGE-008`
- [ ] `RMI-DASHFORGE-010` Unify legacy datasource providers with the sql connector connection layer
  - Depends on: `RMI-DASHFORGE-009`

## Phase 2 — Tier 3: Analytics Provider Contract

**Theme:** dashforgespec (dashboardir + stdlib only) + one generic http connector; omniroadmap as reference provider

- [ ] `RMI-DASHFORGE-011` dashforgespec package: contract types, Go client, JSON Schema generation, SPEC.md, dependency-rule test
- [ ] `RMI-DASHFORGE-012` Provider conformance test-kit (dashforgespec/conformance)
  - Depends on: `RMI-DASHFORGE-011`
- [ ] `RMI-DASHFORGE-013` Generic http connector in the engine via dashforgespec.Client
  - Depends on: `RMI-DASHFORGE-011`
- [ ] `RMI-OMNIROADMAP-023` (repo `github.com/grokify/omniroadmap`) Serve the Analytics Provider contract; conformance-tested; retire the composing-binary requirement
- [ ] `RMI-DASHFORGE-014` End-to-end verification: stock dashforge-server + http source against the reference provider
  - Depends on: `RMI-DASHFORGE-013`

## Phase 3 — Tier 4: Metadata-Table Convention (backlog)

**Theme:** Richer catalogs for SQL-only apps

- [ ] `RMI-DASHFORGE-015` Metadata-table convention: SPEC.md appendix and sql-connector preference logic
  - Depends on: `RMI-DASHFORGE-010`
