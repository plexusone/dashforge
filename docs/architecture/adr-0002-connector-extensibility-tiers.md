# ADR-0002: Connector Extensibility Tiers and the Analytics Provider Contract

**Status:** Accepted — 2026-08-24
**Scope:** dashforge engine, connector-providing applications (omniroadmap, future compass), third-party sources
**Builds on:** ADR-0001 (domain-free engine; connectors are registered by consuming binaries)

## Context

ADR-0001 made the engine connector-free: `analytics.RegisterConnector` is a
compile-time Go registry, and application binaries compose engine + connector
(see `omniroadmap/dashforgeconnector` + `cmd/omniroadmap-server`). That is
correct for first-party integrations, but it cannot scale to DashForge's goal
of connecting to **many sources we do not own**: every source type would need
a Go adapter linked into some binary. Metabase-class tools solve this with
protocol drivers plus runtime-configured sources; DashForge needs the
equivalent, without reintroducing domain dependencies into the engine.

Two mechanisms were considered for DashForge-aware applications to expose
their data: a service API the engine calls, or metadata tables the engine
reads over a database wire. This ADR resolves that as "one neutral contract,
two transports" and defines the full tier model.

## Decision

### Source configuration is always DashForge-side (unchanged)

Sources remain persisted `SourceConfig` records — managed through the UI
panel, the `/api/v1/analytics/sources` API, or hand-edited JSON
(`.dashforge/analytics-sources.json`; ent table when a metadata DB is
configured) — with credentials only ever as OmniVault secret references.
Extensibility adds new values for the `connector` field, never a new
configuration system. `/api/v1/analytics/connectors` reports what a given
deployment supports.

### Four connector tiers

| Tier | Mechanism | Compiled? | For | Owner of adapter |
|------|-----------|-----------|-----|------------------|
| 1 | **Generic protocol drivers** (`sql`, …) | in engine core | databases we don't own but whose protocol is standard (Postgres, MySQL/Dolt, ClickHouse, …) | dashforge |
| 2 | **First-party Go connectors** | composed binary | deep ecosystem integrations (omniroadmap today, compass later) | the app repo |
| 3 | **Analytics Provider contract** (HTTP/JSON) | no — runtime config only | services we don't own, and DashForge-aware apps at arm's length | the service (implements a published spec) |
| 4 | **Metadata-table convention** | no — enriches Tier 1 | SQL-wire-reachable apps with no HTTP surface | the app's schema |

Tier 1 and the Tier 3 generic `http` connector are **domain-free** and belong
in the engine core — they do not violate ADR-0001, which bans *domain*
connectors from core, not protocol drivers.

### Tier 3 is the canonical extensibility mechanism

The **Analytics Provider contract** is a small HTTP/JSON spec:

- `GET  /catalog` → `dashboardir.AnalyticsCatalog`
- `POST /query`  → `dashboardir.AnalyticsQueryRequest` in,
  `dashboardir.AnalyticsQueryResult` out (read-only; the provider enforces
  its own query semantics and limits)
- Auth via bearer/header whose secret lives in the SourceConfig as an
  OmniVault reference; a `version` field in the catalog for schema evolution.

The wire format IS `dashboardir` — the Go types are the source of truth, JSON
Schemas are generated (invopop/jsonschema, linted with schemakit, camelCase
per org convention) and published so non-Go services can implement the
contract. The engine ships one generic `http` connector; a source config is
just `{connector: "http", dsnRef: <base-url-or-secret-ref>, ...}`.

This inverts the dependency: providers do not import dashforge — they
implement a spec. Domain logic (derived fields, app-side query evaluation)
stays in the provider, where it belongs.

**Consequence for omniroadmap:** once it serves `/catalog` + `/query`, stock
`dashforge-server` connects to it with **no composing binary** — Tier 2's
`dashforgeconnector` becomes an optional in-process optimization, and runtime
decoupling matches ADR-0001's code decoupling.

### Tier 4 is an enrichment, not a protocol

Apps reachable only over a database wire may publish agreed tables/views
(e.g. `dashforge_datasets`, `dashforge_fields`) in their own schema; the
Tier 1 `sql` connector prefers them over raw `INFORMATION_SCHEMA`
introspection for display names, roles, and custom-field coverage. Its
limits — SQL-wire only, convention-schema versioning, engine-side query
semantics — are why it is secondary to Tier 3.

## Consequences

- Sources we don't own become reachable three ways without touching Go:
  standard database protocols (Tier 1), the published HTTP contract (Tier 3),
  or metadata-enriched SQL (Tier 4).
- The engine gains exactly two new built-in connectors (`sql`, `http`), both
  domain-free; everything else stays out of core.
- The existing `datasource/` SQL providers (currently dashboard-scoped, never
  wired into the analytics catalog) are the natural starting point for the
  Tier 1 driver — unify rather than duplicate.
- The contract spec needs a home with published JSON Schemas; candidates:
  `docs/specs/` here, or a standalone spec repo (aistandardsio pattern) if
  third parties are expected to implement it.
- Authorization mapping (catalog datasets/fields → SpiceDB resources) is
  transport-independent because it keys off the catalog, not the connector.

## Sequencing

1. Tier 1 `sql` connector in engine core (largest immediate coverage).
2. Analytics Provider contract: schema generation + spec doc + generic `http`
   connector; omniroadmap as reference implementation.
3. Tier 4 table convention when a SQL-only app needs a richer catalog.
4. Compass composes via Tier 2 in-process or consumes Tier 3 unchanged.

## Related

- ADR-0001 — domain-free engine, connector registration at the app binary.
- `omniroadmap/dashforgeconnector`, `omniroadmap/cmd/omniroadmap-server` —
  the Tier 2 reference (RMI-OMNIROADMAP-022).
