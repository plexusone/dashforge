# PLAN — Connector Extensibility — Protocol Drivers and the Analytics Provider Contract

**Initiative:** `INIT-DASHFORGE-002`
**Status:** Draft

## Build Order

### Phase 1 — Tier 1: generic `sql` connector

1. `sql` connector core: driver selection from resolved DSN, catalog via
   `INFORMATION_SCHEMA` introspection (Dolt/MySQL + Postgres), registered as
   `"sql"` in the engine.
2. Query path: dialect `"sql"` execution with read-only allowlist, LIMIT
   clamp, sanitized errors.
3. Unify with the legacy `datasource/` providers: one connection layer, the
   dashboard-scoped feature preserved.

**Exit criteria:** a Dolt database added purely via source configuration is
browsable in the schema browser and queryable end-to-end; engine tests green
against a throwaway dolt sql-server.

### Phase 2 — Tier 3: `dashforgespec` + `http` connector + reference provider

4. `dashforgespec` package: contract/envelope/handshake types, Go client,
   JSON Schema generation (`gen/` + `tools.go`), `SPEC.md`, and the
   dependency-rule test (imports ⊆ {dashboardir, stdlib}).
5. `conformance/` provider test-kit.
6. Generic `http` connector in the engine, delegating to
   `dashforgespec.Client`.
7. Cross-repo: omniroadmap serves `/catalog` + `/query`, passing the
   conformance kit; verify stock `dashforge-server` + an `http` source
   replaces the composing binary.

**Exit criteria:** PRD success criteria for Tier 3; schemas schemakit-lint
clean; conformance kit green against omniroadmap.

### Phase 3 — Tier 4 (backlog)

8. Metadata-table convention as a `SPEC.md` appendix + `sql`-connector
   preference logic. Scheduled only when a SQL-only app needs it.

## Risks and Mitigations

- **Spec churn after extraction** — mitigated by the dependency rule + a
  `ContractVersion` handshake from day one; additive-only within a major.
- **Raw-SQL safety on Tier 1** — statement allowlist and clamps from the
  first commit; no GrokifyQL→SQL compilation promised in this initiative.
- **Double SQL stacks** (datasource/ vs sql connector) — unification is an
  explicit RMI, not an afterthought.
- **Parallel-session collisions** (active repo) — each RMI lands as its own
  scoped commit set; check tree state before starting.

## Follow-On

- `INIT-DASHFORGE-003` — DashForge Cloud & workspace sync (renumbered from
  -002); consumes Tier 3 for hosted-to-customer-provider connectivity.
- Extraction of dashforgespec (+ dashboardir) to a standalone spec repo when
  a third-party implementer appears.
- GrokifyQL→SQL compilation for Tier 1 sources.
