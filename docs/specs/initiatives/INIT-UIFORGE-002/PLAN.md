# PLAN — Persistent Analytics Sources — Metabase-Style Multi-Source Catalog with OmniVault

**Initiative:** `INIT-UIFORGE-002`
**Status:** Draft
**Home repo:** `github.com/plexusone/uiforge`

## Build Order

Three phases; each phase leaves the repo building, tested, and shippable.

### Phase 1 — Source model, secret resolution, persistence

Backend foundation with no behavior change to the running server yet.

1. `SourceConfig` type, validation (including raw-DSN rejection via `omnivault.IsSecretRef`), and the connector registry with the `omniroadmap` factory registered.
2. OmniVault v0.5.0 dependency; server-owned `Resolver` construction with `env://` and `file://` providers.
3. `SourceStore` interface + JSON file store (`.uiforge/analytics-sources.json`), mirroring `questionFileStore`.
4. Ent `AnalyticsSource` schema + codegen + ent-backed store.

**Exit criteria:** `go test ./...` green; both stores round-trip `SourceConfig`; resolver resolves `env://` refs and rejects raw DSNs. No server wiring changed yet.

### Phase 2 — Dynamic service, API, flag removal

5. Rework `analytics.Service`: mutex-guarded provider map; `LoadAll` / `AddSource` / `UpdateSource` / `RemoveSource` / `TestSource`; catalog source-ID re-tagging.
6. CRUD + test + connectors endpoints on `AnalyticsHandler`.
7. Server wiring: resolver → store → service → `LoadAll` in `newServerInternal`; remove `--omniroadmap-dsn`, `Config.OmniRoadmapDSN`, and the hard-coded provider block.

**Exit criteria:** server starts with no source flags; a source POSTed to the API is queryable, survives restart (file-store mode), and DELETE removes it from the catalog. Existing question save/execute policy paths still pass.

### Phase 3 — Builder UI and documentation

8. `AnalyticsSourcePanel` in the Question workspace: list/add/edit/test/remove, catalog refresh on mutation; API client functions in `dashforge.ts`.
9. Docs: update `docs/analytics-catalog.md` (implemented vs. planned credential schemes), README start command, `examples/` config notes.

**Exit criteria:** full walkthrough from the PRD works end-to-end in the browser at `/builder/?mode=questions`; builder lint/typecheck green; docs match the shipped behavior.

## Risks and Mitigations

- **Ent codegen churn** — regenerating `ent/` produces a large diff. Mitigate: isolate codegen in its own commit (`chore`).
- **Startup regression for existing users** — removing `--omniroadmap-dsn` breaks existing start commands. Mitigate: serve a clear error naming the replacement (`unknown flag` is acceptable since Cobra reports it; README and docs updated in the same release), and document the one-time migration: add a source with `dsnRef: env://UIFORGE_OMNIROADMAP_DSN`.
- **Same-connector ID collisions** — two OmniRoadmap sources would emit identical catalog source IDs. Mitigate: service re-tags catalog source ID/Name from `SourceConfig`; covered by a dedicated unit test.
- **Secret leakage via error paths** — dial errors can embed DSNs. Mitigate: sanitize API error responses; log full errors server-side only; test asserts no `dsnRef`-resolved value in API error bodies.
- **SpiceDB resource-ID stability** — analytics authz resource IDs derive from source/dataset IDs. Mitigate: document that `SourceConfig.ID` is immutable after creation (update API rejects ID changes).

## Dependencies

- `github.com/plexusone/omnivault` v0.5.0 (released; verify with `go list -m -versions`).
- `github.com/grokify/omniroadmap` — no changes expected; consumed via existing `store`/`analyticscatalog`/`analyticsquery` packages.
