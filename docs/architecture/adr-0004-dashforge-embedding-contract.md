# ADR-0004: DashForge Embedding Contract — Embed the UI, Not the Tables

**Status:** Accepted — 2026-08-24
**Scope:** dashforge engine, compass (embedding host), any OEM/embedded host
**Builds on:** ADR-0001 (domain-free engine), ADR-0002 (connector tiers), compass ADR-0001 (compass owns the store)

## Context

DashForge has two use cases:

- **Standalone, arms-length** — a Metabase/Looker/Tableau-style end-user app
  that connects to independent data sources.
- **Embedded / OEM** — a GoodData-style engine embedded inside another product,
  giving that product customizable analytics.

For our own use we want the **embedded** case: compass embeds DashForge so
compass gets customizable analytics over its prioritization/portfolio data.
Per compass ADR-0001, compass owns the single DoltDB. The risk in embedding is that the
host stops treating DashForge as an engine with seams and starts treating it as
a pile of tables — reaching directly into DashForge's dashboard/question schema.
That fuses the two products: DashForge can no longer version, upgrade, or ship
standalone independently, and the OEM story collapses.

The requirement, stated plainly: **the integration between dashforge and compass
must be via integration points even when the dashforge UI is embedded in
compass.** Embedding the UI is a rendering detail; it must not merge the data
ownership.

## Decision

DashForge is embedded through **two contracts and nothing else** — never through
shared schema. Owning the *deployment* (one DoltDB, one process) does not mean
owning DashForge's *schema*.

### Contract 1 — data in: the connector API (ADR-0002)

The host exposes its data to DashForge as an analytics **source**, through the
same connector mechanism every other source uses:

- compass registers a **compass connector** (compass ADR-0001) — Tier 2 in-process, or
  Tier 3 over the published HTTP contract — surfacing canonical initiatives plus
  scores/portfolio as a `dashboardir.AnalyticsCatalog`.
- DashForge reads it exactly as it reads any source. No special "embedded" data
  path; the embedded host is just a connector.

### Contract 2 — metadata out: the storage interface

DashForge's analytics **metadata** — saved dashboards, saved Questions, the
data-source registry, chart specs — is reached through a **storage interface**,
not shared tables:

- DashForge defines a metadata-store abstraction (save/load dashboards,
  Questions, source configs).
- **Standalone**: a default DoltDB-backed implementation → DashForge is a
  self-contained Metabase.
- **Embedded in compass**: compass either supplies its own implementation of the
  interface, or points DashForge's default implementation at compass's Dolt
  server as a **separate database**. Either way, compass never reads or writes
  DashForge's schema directly.

This mirrors the seams already in the codebase — the connector registry
(ADR-0002) and the OmniVault `dsnRef` resolver: an interface with a default
implementation, overridable by the host. It is the same move applied to
persistence.

### Deployment ≠ schema ownership

Logical ownership and physical instances are separate concerns. Dolt serves
multiple databases per server, so the embedded compass product can run **one
Dolt server** hosting separate databases — e.g. `compass` (canonical +
prioritization, compass ADR-0001) and `dashforge_meta` (DashForge's metadata store) —
with distinct schemas and owners but a single process to operate. Three logical
stores do not imply three deployments.

## Consequences

- **The same DashForge binary/library serves both use cases** with no fork: the
  standalone app and the embedded engine differ only in which connector is
  registered and which storage implementation is bound.
- **DashForge keeps independent versioning and upgrades** when embedded, because
  compass depends on its interfaces, not its schema.
- **New work: the metadata storage interface.** Today DashForge persists
  dashboards/Questions/source configs directly (`.dashforge/*.json` or its ent
  metadata DB). That direct persistence must be factored behind the storage
  interface so a host can substitute or redirect it. This is the primary
  implementation cost of this ADR.
- **Credentials stay OmniVault references** across the seam (ADR-0002); the host
  supplies secret references, never resolved DSNs, and DashForge never returns
  resolved DSNs/tokens to the browser.
- **Authorization stays catalog-keyed** (ADR-0002): dataset/field → SpiceDB
  resource mapping keys off the catalog, so it is unaffected by whether the
  catalog arrived from a standalone source or an embedding host.

## Related

- ADR-0001 — domain-free engine.
- ADR-0002 — connector tiers; Contract 1 is that mechanism.
- compass ADR-0001 — compass owns the single DoltDB; the connector target and the
  embedding host are compass.
