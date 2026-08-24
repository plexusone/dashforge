# PRD — Connector Extensibility — Protocol Drivers and the Analytics Provider Contract

**Initiative:** `INIT-DASHFORGE-002`
**Status:** Draft
**Home repo:** `github.com/plexusone/dashforge`
**Decision record:** ADR-0002 (connector extensibility tiers)

## Problem

After ADR-0001's decoupling, DashForge's engine ships no connectors; the only
way to add a source type is a compiled Go adapter composed into a binary
(Tier 2 — `omniroadmap/dashforgeconnector` + `cmd/omniroadmap-server`). That
cannot scale to the product goal: connecting to **many sources DashForge does
not own** — ordinary databases, and services that should not require a Go
toolchain or a custom binary to become queryable.

## Goals

- **Tier 1:** a generic `sql` connector in the engine core — any
  Postgres/MySQL/Dolt-wire database becomes a catalog source through
  configuration alone.
- **Tier 3:** the **Analytics Provider contract** — a published HTTP/JSON
  spec (`GET /catalog`, `POST /query` speaking `dashboardir` types) consumed
  by one generic `http` connector, so any service implementing two endpoints
  becomes a source at runtime. omniroadmap is the reference provider,
  after which stock `dashforge-server` reaches it with no composing binary.
- **Extractable spec:** the contract is authored as the `dashforgespec`
  package inside this repo, importing **only `dashboardir` and the standard
  library** — `dashboardir` is explicitly the consumer-facing interface
  (informal: plain types, not Go interfaces). The future extraction unit is
  dashforgespec + dashboardir together; moving them out is a path change,
  not a refactor.
- Source *configuration* stays exactly as built: UI panel, REST API, or JSON
  file, with OmniVault secret references.

## Non-Goals

- Tier 4 (metadata-table enrichment of the sql tier) — backlogged here,
  implemented when a SQL-only app needs a richer catalog.
- Out-of-process Go plugins — the HTTP contract covers the need.
- Write/mutation support — providers are read-only analytics sources.
- Standing up a separate spec repository now — deferred until a third party
  wants to implement the contract (extraction-ready by design).
- Cloud/workspace sync — renumbered to `INIT-DASHFORGE-003`; Tier 3 is its
  foundation (hosted DashForge querying customer-side providers).

## Users and Experience

- **Operator with a database:** adds a source in the UI — connector `sql`,
  a `dsnRef` — and the schema browser fills with tables/columns; Questions
  and dashboards work with no code anywhere.
- **Service owner (any language):** implements `/catalog` + `/query` against
  the published JSON Schemas; the operator registers it with connector
  `http` and a base URL. No Go, no DashForge imports.
- **omniroadmap user:** runs stock `dashforge-server`; omniroadmap serves
  the contract itself. `omniroadmap-server` remains as an in-process
  optimization, no longer a requirement.

## Success Criteria

- A Dolt/Postgres database is browsable and queryable end-to-end via a
  `sql`-connector source added purely through configuration.
- `dashforgespec` compiles with imports limited to `dashboardir` + stdlib
  (enforced by a test), with generated, schemakit-linted JSON Schemas.
- omniroadmap serves the contract; stock `dashforge-server` (no composing
  binary) lists, catalogs, and queries it via the `http` connector.
- `/api/v1/analytics/connectors` reports `sql` and `http` in stock builds.
