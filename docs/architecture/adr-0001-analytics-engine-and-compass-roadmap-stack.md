# ADR-0001: Generic Analytics Engine and the Compass Roadmap/Prioritization Stack

**Status:** Accepted — 2026-08-22
**Scope:** ecosystem (dashforge, omniroadmap, prism-roadmap, prism-core, compass-rice, compass)
**Home note:** authored in the dashforge repo (the near-term actionable outcome is dashforge's
connector decoupling); mirror into `ProductBuildersHQ/compass` when that repo exists.

## Context

dashforge (renamed from uiforge) is meant to be a **generic analytics engine** — Metabase/
Tableau/Hex-class: catalogs, Questions, dashboards over pluggable data sources. But dashforge
core currently compiles a specific data source into itself: `internal/server/analytics/
omniroadmap.go` registers the OmniRoadmap connector from inside the engine, pulling omniroadmap
and its PM-tool integrations (aha-go, productboard-go, go-atlassian, prism-roadmap, aha-studio,
omniroadmap-core) transitively into a supposedly domain-neutral engine. That is the coupling
this ADR removes, and it clarifies where the roadmap/prioritization domain actually lives.

## Decision

### Layering (acyclic)

```
prism-core                                     foundation primitives
   └─ prism-roadmap                            canonical planning/roadmap MODEL + generic frameworks
aha-go / productboard-go / go-atlassian
   └─ omniroadmap                              ETL from providers + Dolt store (uses prism-roadmap types)
compass-rice  (may import prism-roadmap)       advanced RICE: domain investment profiles
dashforge                                      generic analytics engine (NO domain deps)
   ▼
compass  (ProductBuildersHQ)                   the app: composes omniroadmap + dashforge + compass-rice
                                               + prism-roadmap frameworks into a branded roadmap/
                                               prioritization surface
```

No cycles: prism-roadmap depends on nothing prioritization- or app-specific; compass is the
sole convergence point.

### dashforge stays domain-free

- Promote the connector contract (`CatalogProvider`, `QueryProvider`, `RegisterConnector`, the
  catalog IR) out of `internal/` into a public package (e.g. `dashforge/analytics` or
  `dashforge/connector`) so external apps can register connectors.
- Move `omniroadmap.go` **out** of dashforge core onto the consumer side.
- Compose at the app binary: `compass` imports dashforge (engine) + omniroadmap (data) and
  registers the OmniRoadmap connector. `dashforge-server` remains a generic engine.

### Framework home rule: generic in prism-roadmap; promote to `compass-*` on domain extension

- **prism-roadmap** owns generic/textbook framework definitions: MoSCoW, Kano, and (until
  consolidated) RICE, plus rubrics, goals, effort, journey.
- A framework graduates to a `compass-*` module **only when it grows domain-specific extensions
  beyond the generic definition** — the RICE → compass-rice precedent (compass-rice adds six
  normalized investment profiles: Customer, Platform, Market Expansion, Operational Efficiency,
  Supportability, Risk).
- **RICE consolidation:** generic Intercom-style RICE becomes a compass-rice *profile* (same Go
  structs, flagged as the simple/alternative rating vs. the domain profiles). RICE is then
  **removed from prism-roadmap** — prism-roadmap does NOT import compass-rice (dependency arrow
  stays compass-rice → prism-roadmap, never the reverse). Owned by the compass-rice agent.
- **MoSCoW and Kano stay in prism-roadmap** as generic base frameworks; compass consumes them.
  They promote only if they gain compass-rice-style specialization.

### Canonical roadmap item struct: core in prism-roadmap, composite in compass

Canonical domain types must live where every layer can depend on them — a library, never the
app (otherwise the data layer would have to import the app, inverting dependencies).

- **Core canonical roadmap item → prism-roadmap.** Identity, title, status, provider source
  refs, hierarchy, dates, and a generic custom-field/evidence bag. **Prioritization-agnostic** —
  no compass-rice fields embedded. omniroadmap ingests providers into this type and maps its
  Dolt/ent store to/from it.
- **Prioritization → compass-rice**, composed on top of the core item, not embedded in it.
- **Full item (core + scores + denormalized analytics fields) → compass** as an app view model,
  and/or materialized into omniroadmap's store when compass writes computed scores back for query
  performance. This is a composition of the two library types, not a canonical struct others import.

### Prioritization scoring is a decision/analytics concern (compass), not data (omniroadmap)

omniroadmap stores raw evidence (reach, ARR, accounts, SAM/SOM, effort, custom fields). compass
computes scores (compass-rice profiles, MoSCoW, Kano, TAM/SAM/SOM) and drives analytics via
dashforge. omniroadmap does not import compass-rice; compass may write computed scores back to
omniroadmap as materialized metadata if analytics performance requires it.

## Consequences

- dashforge sheds seven domain dependencies by moving one file and promoting the connector
  contract; `dashforge-server` becomes a truly generic engine.
- omniroadmap stays a framework-neutral roadmap data/ETL library, reusable by any consumer.
- prism-roadmap remains a broadly-reusable model library, not coupled to a ProductBuildersHQ
  prioritization product.
- compass is the only place the full stack converges — the branded roadmap builder/prioritization
  surface.
- One prioritization model (prism-roadmap base frameworks + compass-rice advanced profiles),
  not two divergent RICE implementations.

## Related work

- dashforge rename & split: `INIT-DASHFORGE-001` (dashforge decoupling belongs here or as a
  precursor so dashforge is born decoupled).
- compass-rice: absorb generic RICE as a profile (compass-rice agent).
- prism-roadmap: drop RICE, own the canonical roadmap item struct.
- New: `INIT-COMPASS-001` (ProductBuildersHQ/compass) composing the stack.
