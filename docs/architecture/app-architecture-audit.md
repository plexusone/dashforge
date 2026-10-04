# DashForge Application-Architecture Audit

**Date:** 2026-09-11
**Tracking:** RMI-DASHFORGE-019
**Audited against:** SystemForge Application Architecture (Normative convention v1)

This document maps DashForge's **current** package structure onto the
SystemForge three-layer convention (`internal/foundation`,
`internal/platform`, `internal/feature/<slug>`, public `app/`). It is an
analysis artifact only — no code is moved by this RMI. The dual-wiring pilot is
scoped separately as RMI-DASHFORGE-020.

## Convention summary (what we are grading against)

- Three layers under `internal/`: `foundation/` (identity/authn/authz/sessions,
  always a single shared system of record), `platform/` (db, observability,
  config, jobs, events — reusable infra, per-deployment), and
  `feature/<slug>/` (product capability verticals).
- Feature slice files are conventional (not mandatory): `feature.go`,
  `contract.go`, `routes.go`, `handler.go`, `service.go`, `repository.go`,
  `model.go`, `client.go`.
- Dependency rule: `feature → platform → foundation` allowed; the reverse
  denied; `feature A → feature B` only through B's `contract.go`.
- A **public** `app/` package (not `internal/app`) exports
  `New(cfg) (*App, error)` where `*App` implements the composition-runtime
  `Feature` interface — the seam other repos import.
- One canonical feature slug drives Go package / UI route / BFF path / REST
  path / permission prefix / telemetry.

## Current shape (starting point)

DashForge today is a set of flat top-level Go packages (the JSON-IR library
surface plus product capabilities), an `ent/` generated ORM package, an
`internal/` tree containing only `authz/` and `server/`, two `cmd/` binaries,
and TypeScript/JS asset directories (`renderer/`, `ts/`, `cube/`, and the
embedded `viewer/` / `builder/`). Composition happens in
`internal/server/server.go` (`server.New`), with a second composition path in
`multiapp/backend.go` (`server.NewServerWithDatabase`). There is **no `app/`
package** and no `compose.Feature` seam yet.

## Package → target mapping

> **Superseded in part by the uiforge extraction (RMI-DASHFORGE-007):** `uispec/`, `registry/`, `pkg/{diff,expression,interaction,state}/`, the page/component halves of `schema/`, and the TS `renderer/` now live in `github.com/plexusone/uiforge`, consumed as a module dependency (v0.1.0+). Rows below describing those packages are historical. `bridge/` remains in-repo as the DashboardIR → UISpec adapter, and `schema/` retains only the dashboard schema.

| Current package | Target | Slug | Justification |
| --- | --- | --- | --- |
| `uispec/` | **Public library (IR)** — keep at module root | — | Canonical UISpec JSON type system (`PageSpec`, `ComponentInstance`, layout, theme, bindings). The product's public SDK surface; imported by `renderer`, `builder`, `bridge`, and external consumers. Shared domain vocabulary, not a vertical. |
| `dashboardir/` | **Public library (IR)** — keep, but split | — | Dashboard IR types **plus** analytics catalog/query DTOs (`AnalyticsCatalog`, `AnalyticsQueryRequest/Result`, `SavedQuestion`). Dashboard IR stays public; the analytics query/catalog DTOs are the wire types for the analytics contract and should be split toward it (see risks). |
| `bridge/` | Public util (or fold into `feature/dashboard`) | — | Single function `DashboardToPageSpec` converting legacy DashboardIR → UISpec. Trivial; can stay a public adapter or move beside the dashboard feature. |
| `pkg/diff`, `pkg/expression`, `pkg/interaction`, `pkg/state` | **Public library** — keep under `pkg/` | — | Generic UISpec runtime utilities (structural diff, `${…}` expression eval, event→action engine, page state store). Already correctly placed as reusable library code. |
| `registry/` | **Public library** — keep | — | Component manifest registry + profile constraints over UISpec. Library surface used to validate/compose specs. |
| `schema/` | **Public library** — keep | — | `//go:embed` JSON schemas generated from Go types. Library asset. |
| `viewer/`, `builder/` | **Public** embed packages — keep | — | `embed.FS` wrappers around built front-ends, served by feature handlers. Thin, no business logic. |
| `renderer/`, `ts/`, `cube/` | **Non-Go** — leave in place | — | TS renderer, TS frontend sources, and Cube.js semantic-layer service dir. Out of scope for the Go layering; `cube/` intentionally has no Go/lint tooling. |
| `analytics/` | **Split:** SPI stays public + `internal/feature/analytics` | `analytics` | `Service`/`SourceStore`/`resolver`/connector-registry orchestration → feature. But `CatalogProvider`/`QueryProvider` are an inbound SPI that external repos implement; that interface **must stay public/importable** (see risks). |
| `connectors/sqlsource/` | `internal/platform/datasource/...` (plugin impl) | — | Concrete Tier-1 generic SQL connector registered via `init()`. Infra plugin under the datasource capability. |
| `datasource/` | `internal/platform/datasource` (with feature-facing CRUD) | `datasource` | Provider plugin architecture for external DB connections (`Provider`, `Connection`, `Manager`, `QueryExecutor`). Primarily reusable infra → platform; the datasource CRUD REST surface is the thin feature layer over it (dual nature — see risks). |
| `multiapp/` | **Replace** with public `app/` | — | `Backend` already implements SystemForge's `multiapp.AppBackend` (`Slug()="dashforge"`, `Routes(deps)`, lifecycle). This is the raw material for `app.New` implementing `compose.Feature`; the canonical app slug `dashforge` is already established here. |
| `ent/` | `internal/platform/db` (shared client) — **spans layers** | — | Single generated ORM package/client covering foundation entities (principal, organization, membership, human, oauth_account, refresh_token, user) **and** feature entities (dashboard*, saved_query, analytics_source, datasource, alert*, integration, marketplace: license/listing/publisher/seat_assignment/subscription). Cannot be split cleanly along layer lines (see risks). |
| `internal/authz/` | `internal/foundation/authz` | — | SpiceDB-backed authorization `Service`, modes, schema. Foundation layer (shared system of record). |
| `internal/server/auth/` | `internal/foundation/authn` (+ sessions) | — | JWT service + OAuth handler (GitHub/Google/SystemAuth). Identity/authn → foundation. |
| `internal/server/db/` | `internal/platform/db` | — | `Database` interface + `Open`. Reusable infra. |
| `internal/server/config/` | `internal/platform/config` | — | Config loading. Reusable infra. |
| `internal/server/middleware/` | `internal/platform/http` (middleware) | — | Cross-cutting HTTP middleware. |
| `internal/server/server.go` | **Dissolve** into `app/` + platform + feature routes | — | Current composition root (`New`, `newServerInternal`, `setupRoutes`). Wiring logic → `app.New`; per-domain route registration → each feature's `routes.go`/`handler.go`. |
| `internal/server/api/api.go` (dashboard CRUD) | `internal/feature/dashboard` | `dashboard` | `Handler` = dashboard list/get/create/update/delete + query. Owns dashboard, dashboard_template, dashboard_version entities. |
| `internal/server/api/question.go` + `question_huma.go` + `grokifyql_policy.go` | `internal/feature/query` | `query` | Saved-question (GrokifyQL) CRUD + Huma routes + policy provider. Owns `saved_query`. Publishes app capability `dashforge.question.write`. `GrokifyQLPolicyProvider` is already an interface. |
| `internal/server/api/analytics.go` + `analytics_sources.go` | `internal/feature/analytics` | `analytics` | Analytics catalog + query-execution + source-config endpoints, over the public `analytics.Service`. Publishes `dashforge.query.run`, `dashforge.field_values.read`. |
| `internal/server/api/datasource.go` | `internal/feature/datasource` (handler) | `datasource` | REST CRUD over `datasource.Manager`/`QueryExecutor`. Thin feature layer atop the platform datasource infra. |
| `internal/server/api/alert.go` + `integration/alert/` | `internal/feature/alert` | `alert` | Alert CRUD/enable/disable/events + the evaluation `Engine` (threshold/schedule/data-change evaluators). Owns alert, alert_channel, alert_event. |
| `internal/server/api/integration.go` + `integration/channel/*` | `internal/feature/integration` | `integration` | Integration CRUD + notification `Channel` plugin adapters (email/slack/webhook/whatsapp). Owns `integration`. |
| `internal/server/api/marketplace.go` | `internal/feature/marketplace` | `marketplace` | Template marketplace CRUD. Owns publisher, listing, license, subscription, seat_assignment, dashboard_template. |
| `internal/server/api/ai.go` | `internal/feature/assistant` | `assistant` | AI dashboard-generation handler (provider fallback, Ollama). API-only feature. |
| `cmd/dashforge/` | `cmd/dashforge` (static viewer dev server) | — | Cobra CLI: `serve` runs a plain file server for the embedded viewer + JSON files. Keep as a deployment entrypoint; does **not** wire the full app. |
| `cmd/dashforge-server/` | `cmd/dashforge` (monolith) via `app.New` | — | Cobra CLI wrapping `internal/server.New`. Should call `app.New(cfg)` once `app/` exists. |
| `examples/`, `testdata/`, `docs/`, `site/` | Unchanged | — | Fixtures, docs, generated site. |

## Assigned canonical feature slugs

| Slug | Capability | Primary current source | Owning entities |
| --- | --- | --- | --- |
| `dashboard` | Dashboard CRUD + IR storage | `internal/server/api/api.go` | dashboard, dashboard_template, dashboard_version |
| `query` | Saved GrokifyQL questions + policy | `api/question*.go`, `grokifyql_policy.go` | saved_query |
| `analytics` | Catalog + ad-hoc query execution + source config | `analytics/`, `api/analytics*.go` | analytics_source |
| `datasource` | External DB connections CRUD | `datasource/`, `api/datasource.go` | datasource |
| `alert` | Alert engine + CRUD | `integration/alert/`, `api/alert.go` | alert, alert_channel, alert_event |
| `integration` | Notification channels + CRUD | `integration/channel/`, `api/integration.go` | integration |
| `marketplace` | Template marketplace | `api/marketplace.go` | publisher, listing, license, subscription, seat_assignment |
| `assistant` | AI dashboard generation | `api/ai.go` | — (API-only) |

The application (composition boundary) slug is **`dashforge`** — already the
value returned by `multiapp.Backend.Slug()`.

## Contracts (`contract.go`) to extract

| Feature | Interface (suggested) | Key methods | Status today |
| --- | --- | --- | --- |
| `analytics` | `analytics.QueryProvider` (embeds `CatalogProvider`) | `Catalog(ctx)`, `Query(ctx, AnalyticsQueryRequest) (AnalyticsQueryResult, error)`, `Close()` | **Already exists and exported.** Best-formed contract in the repo; already implemented by `*analytics.Service` and (by design) by external provider repos. |
| `query` | `query.PolicyProvider` (exists as `GrokifyQLPolicyProvider`) + a `query.Store` | `Policy(ctx, sourceID)`; question CRUD | Policy interface exists in `api/grokifyql_policy.go`; question store is currently concrete (`SavedQuestionHandler`). |
| `datasource` | `datasource.Executor` | `Execute(ctx, QueryRequest) (*QueryResult, error)`, `GetSchema(ctx, SchemaRequest)` | Concrete today (`*datasource.QueryExecutor`); interface is a small, obvious extraction. |
| `dashboard` | `dashboard.Store` | `List/Get/Create/Update/Delete(ctx, …)` | Concrete `Handler` over ent; extract for cross-feature reuse (e.g. alerts referencing dashboards). |
| `alert` | `alert.Evaluator` (exists) / `alert.Engine` | `Evaluate(ctx, …)` | `Evaluator`/`Engine` exist as concrete types; interface extraction is minor. |
| `integration` | `channel.Channel` (exists) | send/notify | Already an interface (`channel.Channel`) with plugin adapters. |

## The public `app/` package

**What `cmd/dashforge-server` does today:** `runServe` reads flags into a
`server.Config`, builds a `*slog.Logger`, calls `server.New(cfg, logger)`, and
calls `srv.ListenAndServe()`. All real wiring is inside
`internal/server.newServerInternal`: open DB + migrate/RLS, build the
datasource `Manager`/`QueryExecutor`, build the analytics `Service` (ent or
file source store + OmniVault resolver), build JWT/OAuth, build the authz
`Service` (simple or SpiceDB), build the GrokifyQL policy provider, build the
saved-question and AI handlers, then `setupRoutes` mounts every handler on a
chi mux fronted by Huma.

**What moves into `app.New(cfg) (*App, error)`:** the entire
`newServerInternal` assembly, restructured so each capability is constructed as
a feature slice and registered via the SystemForge `compose.Feature` seam
(`Name()`, `RegisterRoutes(Registrar)`, `Start`, `Stop`). `*App` aggregates the
features and itself implements `Feature`, so `cmd/dashforge` (monolith) and any
embedding repo compose it identically. The two existing composition paths
(`server.New` standalone and `multiapp.Backend` via `NewServerWithDatabase`)
collapse into this one public entrypoint; `multiapp.Backend` becomes a thin
adapter over `app.New` (or is retired). Platform primitives (db, config,
authz/authn) are constructed once and injected into features as dependencies.

## Recommended dual-wiring pilot (RMI-DASHFORGE-020)

**Pilot capability: `analytics` query execution, via the existing
`analytics.QueryProvider` contract.** (The hypothesis named `query` or
`datasource`; validated against reality, `analytics` query execution is the
strongest, with `datasource` a clean second.)

Why `analytics` wins:

1. **The contract already exists and is exported** — `QueryProvider` /
   `CatalogProvider` in `analytics/catalog.go`. No interface needs inventing.
2. **The wire DTOs already exist and are JSON-serializable** —
   `dashboardir.AnalyticsQueryRequest` / `AnalyticsQueryResult` /
   `AnalyticsCatalog`. A remote `Client` marshals these directly.
3. **A REST endpoint already exists** — `/api/v1/analytics/query` and the
   catalog endpoints in `api/analytics.go`. The remote `Client` has a concrete
   wire target to adapt to.
4. **It is already designed for pluggable providers** — the doc comment states
   external repos implement `CatalogProvider` without depending on DashForge
   storage, so the local/remote symmetry is inherent to the design.

The local implementation is the existing `*analytics.Service`; the remote
`client.go` implements the same `QueryProvider` interface by calling the
analytics REST endpoints. Consumers (the analytics handler, dashboard widgets)
depend only on the interface, so only construction changes between topologies.

**Secondary candidate: `datasource`.** `QueryExecutor.Execute(ctx,
QueryRequest) (*QueryResult, error)` is a narrow, obvious contract with an
easy remote client, but it is lower-level platform infra rather than a product
capability, and `QueryResult` carries DB-shaped types that need a wire-friendly
DTO pass first.

## Migration risk & ordering

**Mechanical (low risk, do first):**

- `internal/server/db/` → `internal/platform/db`
- `internal/server/config/` → `internal/platform/config`
- `internal/server/middleware/` → `internal/platform/http`
- `internal/authz/` → `internal/foundation/authz`
- `internal/server/auth/` → `internal/foundation/authn`
- `schema/` (dashboard-only), `viewer/`, `builder/` stay put. (`pkg/*`, `uispec/`, `registry/` have moved to uiforge.)

**Tangled (higher risk, sequence carefully):**

- Splitting `internal/server/api/*.go` into per-feature `handler.go`/`routes.go`
  — every handler currently shares the `New*(database db.Database, logger)`
  pattern and is mounted directly on one chi mux; the Huma vs. chi split
  (saved questions use Huma; the rest use chi) must be reconciled behind the
  `Registrar` abstraction.
- Unifying `server.New` + `multiapp.Backend` into `app.New` — two composition
  paths with different DB setup (standalone open vs. schema-scoped pool) must
  merge without regressing multi-app schema isolation.
- Introducing the `compose.Feature` seam depends on the SystemForge
  composition-runtime package being available to import.

## Does-not-fit-cleanly notes

1. **`ent/` is one generated package spanning foundation and feature
   entities.** Identity entities (principal, organization, membership, human,
   oauth_account, refresh_token, user) belong to `foundation`; the rest belong
   to features. Ent generates a **single** client and package, so the
   convention's per-feature `repository.go` boundary and the
   `foundation`/`feature` separation cannot both be honored by physically
   splitting `ent/`. Practical resolution: keep one shared ent client in
   `internal/platform/db`, and let each feature's `repository.go` wrap it with
   feature-scoped queries — accepting that the generated model layer is shared
   rather than layer-partitioned. Foundation-owned entities are the sharpest
   edge (a shared client technically lets a feature touch identity tables
   directly, which review must guard).

2. **The `analytics` provider SPI must stay public, but the convention puts
   feature code under `internal/`.** `CatalogProvider`/`QueryProvider` are an
   inbound plugin interface implemented by external consumer repos; Go
   `internal/` visibility would make them unimportable cross-module. So the
   analytics contract types must live in a public package (module root, e.g.
   the existing `analytics` package or the `dashboardir` DTOs) even though the
   analytics *orchestration* moves to `internal/feature/analytics`. This is the
   one place where "feature" and "public library surface" legitimately overlap.

3. **`datasource` is genuinely dual-natured** — reusable infra (platform) and
   a product-facing CRUD capability (feature). The split here (platform engine
   + thin feature handler over its `contract.go`) is a judgment call, not a
   clean single-bucket assignment.

4. **Two `cmd/` binaries do very different things.** `cmd/dashforge serve` is a
   static viewer/file dev server (no DB, no API), while `cmd/dashforge-server
   serve` is the full application. They should not both be treated as
   "the monolith entrypoint"; only `dashforge-server` maps to `app.New`.
</content>
</invoke>
