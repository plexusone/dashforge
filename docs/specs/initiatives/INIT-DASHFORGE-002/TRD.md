# TRD — Connector Extensibility — Protocol Drivers and the Analytics Provider Contract

**Initiative:** `INIT-DASHFORGE-002`
**Status:** Draft

## Current State

- `analytics.RegisterConnector(name, factory)` is the compile-time registry;
  `SourceConfig` persistence, OmniVault dsnRef resolution, catalog
  re-tagging, and the sources CRUD/UI all exist and are connector-agnostic.
- `dashboardir` holds the neutral analytics types (`AnalyticsCatalog`,
  `AnalyticsQueryRequest`, `AnalyticsQueryResult`) — explicitly the
  consumer-facing interface for outside consumers (informal: plain structs,
  camelCase JSON tags, no Go interfaces).
- `datasource/` contains legacy SQL providers (postgres, mysql) with their
  own manager/executor, used for dashboard-scoped bindings — never wired
  into the analytics catalog.
- Tier 2 reference exists: `omniroadmap/dashforgeconnector`.

## Design

### Tier 1 — generic `sql` connector (engine core)

A built-in connector registered as `"sql"`:

- **Dial:** the resolved DSN selects the driver by scheme/form (reuse the
  `mysql://`/`postgres://` handling patterns from `internal/server/db`).
- **Catalog:** introspect `INFORMATION_SCHEMA` (tables → datasets, columns →
  fields with type mapping to `dashboardir` field types; `filterable`/
  `sortable`/`selectable` true by default). Dataset `queryName` = table name.
- **Query:** dialect `"sql"` — single-statement, read-only enforcement
  (statement-type allowlist: SELECT only), server-side LIMIT clamp using the
  engine's existing max-rows discipline, parameter passthrough. GrokifyQL
  over SQL sources is *not* in scope for this initiative (a GrokifyQL→SQL
  compiler is future work; the existing GrokifyQL policy path applies only
  to dialects that declare it).
- **Unification:** `datasource/` providers' connection handling is folded
  into (or reused by) this connector so there is one SQL connection layer;
  the dashboard-scoped `DataSource` feature keeps its API surface but stops
  duplicating pooling/dial logic.

### Tier 3 — `dashforgespec` package (the Analytics Provider contract)

Location: `dashforgespec/` at the repo root.

**Dependency rule (enforced by a test):** imports limited to
`github.com/plexusone/dashforge/dashboardir` and the standard library.
`dashboardir` is the wire format — no parallel type set, no mapping layer.
Extraction unit is dashforgespec + dashboardir together; a future
`dashforge-spec` repo move is an import-path change only. Schema generation
lives in `dashforgespec/gen/` behind `//go:build ignore` with a same-dir
`tools.go` (per org convention) so generator deps (invopop/jsonschema) never
become package deps.

Contents:

- `contract.go` — endpoint constants (`/catalog`, `/query`), the version
  constant, auth header name; `ErrorEnvelope{Code, Message}`; a
  `ProviderInfo{Name, ContractVersion}` handshake type returned in the
  catalog response envelope.
- `client.go` — a small Go client (stdlib `net/http`) implementing
  `Catalog(ctx)` / `Query(ctx, req)` against a base URL + bearer token.
  This is what the engine's `http` connector uses, and providers can use it
  in their own tests.
- `conformance/` — a provider test-kit: given a base URL, exercises catalog
  shape, query round-trip, read-only rejection, and error-envelope format.
- `schema/` — generated JSON Schemas for the wire types (dashboardir types +
  envelope), `//go:embed`-ed, schemakit-linted with `--property-case
  camelCase` (API/document surface per org convention).
- `SPEC.md` — the normative document: endpoints, auth (bearer token; secret
  handled DashForge-side as a dsnRef-style reference), versioning
  (`ContractVersion` in the handshake; additive-only within a major),
  read-only semantics, limits, and error model.

### Tier 3 — generic `http` connector (engine core)

Registered as `"http"`. Factory input (the resolved "DSN") is the provider
base URL, optionally with an embedded bearer-token reference resolved like
any dsnRef. Delegates to `dashforgespec.Client`; `Catalog`/`Query` pass
`dashboardir` types straight through. Timeouts, response-size caps, and
sanitized errors follow the engine's existing connector discipline.

### Reference provider — omniroadmap (cross-repo)

omniroadmap mounts the contract (its existing `analyticscatalog`/
`analyticsquery` behind `/catalog` + `/query`, using
`dashforgespec/conformance` in tests). Once serving, stock
`dashforge-server` + an `http` source replaces the composing-binary
requirement; `cmd/omniroadmap-server` stays as an optimization.

### Tier 4 — metadata-table convention (backlog)

Agreed tables/views (`dashforge_datasets`, `dashforge_fields`) that the
`sql` connector prefers over raw introspection. Specified in `SPEC.md` as an
appendix when implemented.

## Security Considerations

- `sql` connector: read-only statement allowlist, LIMIT clamp, no
  multi-statement, sanitized driver errors (existing discipline).
- `http` connector: outbound only, TLS-by-default guidance in SPEC.md,
  bearer secrets as OmniVault references, response-size and time bounds.
- Authorization mapping is transport-independent — SpiceDB resources key off
  the catalog, not the connector.

## Testing Strategy

- dashforgespec: dependency-rule test (parse imports of the package and fail
  on anything outside dashboardir/stdlib); schema-generation golden test;
  client round-trip against `httptest` servers; conformance kit self-test.
- sql connector: introspection + query against throwaway Dolt (godolt
  harness pattern, skip-if-absent) and sqlite/postgres where available.
- http connector: engine-level test with an `httptest` provider implementing
  the contract; end-to-end: sources CRUD → catalog merge → query.
- omniroadmap provider: conformance kit in its own CI.
