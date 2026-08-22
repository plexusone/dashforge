# TRD — Persistent Analytics Sources — Metabase-Style Multi-Source Catalog with OmniVault

**Initiative:** `INIT-UIFORGE-002`
**Status:** Draft
**Home repo:** `github.com/plexusone/uiforge`

## Current State

- `internal/server/analytics.Service` holds an immutable `[]CatalogProvider` built once in `newServerInternal` (`internal/server/server.go`); the only provider is `OmniRoadmapProvider`, constructed from `Config.OmniRoadmapDSN`, which comes solely from the `--omniroadmap-dsn` CLI flag.
- `docs/analytics-catalog.md` already specifies the target model: persisted source metadata with `dsnRef` secret references resolved at runtime via OmniVault, never returning resolved DSNs to the browser.
- `SavedQuestionHandler` (`internal/server/api/question.go`) establishes the persistence pattern to mirror: metadata DB when configured, JSON file fallback (`.uiforge/questions.json`) otherwise.
- The generic `datasource.Manager` / ent `DataSource` path is for raw SQL connections and is out of scope; analytics sources are application-catalog connectors.

## Design

### Source configuration model

New types in `internal/server/analytics` (JSON camelCase, Go structs are source of truth):

```go
// SourceConfig describes one persisted analytics source.
type SourceConfig struct {
    ID        string    `json:"id"`        // slug, unique, e.g. "omniroadmap-local"
    Name      string    `json:"name"`      // display name
    Connector string    `json:"connector"` // registry key, e.g. "omniroadmap"
    DSNRef    string    `json:"dsnRef"`    // secret reference, e.g. env://UIFORGE_OMNIROADMAP_DSN
    Enabled   bool      `json:"enabled"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

`DSNRef` must be a valid OmniVault secret reference (`omnivault.IsSecretRef`); raw DSNs are rejected at validation time so they can never be persisted.

### Connector registry

```go
// ConnectorFactory opens a QueryProvider from a resolved DSN.
type ConnectorFactory func(dsn string) (QueryProvider, error)
```

Package-level registry `map[string]ConnectorFactory`; `"omniroadmap"` registers the existing `NewOmniRoadmapProvider`. Adding a future connector (VisionStudio, etc.) is one `Register` call. The registry also feeds the API's connector-list endpoint so the UI can render a dropdown.

### Secret resolution — OmniVault v0.5.0

`github.com/plexusone/omnivault` (v0.5.0) `Resolver` is constructed once per server:

- Registered schemes in this pass: `env://` (`providers/env`) and `file://` (`providers/file`).
- Architecture uses the full `Resolver` so `keyring://`, `aws-sm://`, and `sql://` providers can be registered later without redesign.
- Resolution happens only inside connect/test paths; the resolved DSN lives in memory for the duration of the dial and is never logged, persisted, or serialized.
- Tests exercise `env://` only.

### Persistence — store interface with two backends

```go
type SourceStore interface {
    List(ctx context.Context) ([]SourceConfig, error)
    Get(ctx context.Context, id string) (SourceConfig, error)
    Save(ctx context.Context, cfg SourceConfig) (SourceConfig, error) // create or update
    Delete(ctx context.Context, id string) error
}
```

- **Ent store** — new ent schema `AnalyticsSource` (fields mirroring `SourceConfig`; `dsn_ref` is a plain string — it is a reference, not a secret). Used when the metadata DB is configured. No organization edge in this pass.
- **File store** — `.uiforge/analytics-sources.json` (path configurable via `--analytics-source-store`, defaulting alongside the question store). Mutex-guarded read-modify-write, same as `questionFileStore`. Safe on disk because it contains only references.

Selection logic in `newServerInternal`: ent store if `db != nil`, else file store.

### Dynamic analytics service

`analytics.Service` becomes mutable:

```go
type Service struct {
    mu        sync.RWMutex
    resolver  *omnivault.Resolver
    store     SourceStore
    providers map[string]QueryProvider // keyed by SourceConfig.ID
}
```

- `LoadAll(ctx)` — called at startup; connects every enabled stored source. A source that fails to connect logs a warning and is reported as errored, not fatal (server still starts).
- `AddSource(ctx, cfg)` — validate, resolve, connect, persist, register.
- `UpdateSource(ctx, cfg)` — persist, close old provider, reconnect.
- `RemoveSource(ctx, id)` — close provider, delete from store.
- `TestSource(ctx, cfg)` — resolve + connect + catalog fetch, then close; nothing persisted.
- `Catalog(ctx)` / `Query(ctx, req)` — as today, iterating the provider map under `RLock`. Source IDs in the combined catalog are the `SourceConfig.ID` values, so provider catalogs must be re-tagged if their internal source ID differs (the OmniRoadmap catalog builder emits `omniroadmap`; the service overrides `Sources[i].ID`/`Name` with the config's ID/Name to keep IDs stable and unique across multiple instances of the same connector).

`HasProviders`, `Close`, and the `GrokifyQLPolicyProvider` integration keep their current semantics.

### HTTP API

Extend `AnalyticsHandler` (`internal/server/api/analytics.go`):

| Method | Path | Behavior |
|--------|------|----------|
| GET | `/api/v1/analytics/sources` | List `SourceConfig`s plus runtime status (`connected`/`error`/`disabled`) |
| POST | `/api/v1/analytics/sources` | Create + connect |
| PUT | `/api/v1/analytics/sources/{id}` | Update + reconnect |
| DELETE | `/api/v1/analytics/sources/{id}` | Disconnect + delete |
| POST | `/api/v1/analytics/sources/test` | Test a candidate config (unsaved) |
| GET | `/api/v1/analytics/connectors` | List registered connector types |

Responses carry `dsnRef` verbatim (it is a pointer, not a secret) but never any resolved value. Error messages from failed dials are sanitized: connector errors are logged server-side; the API returns a generic failure plus the vault-scheme error class (e.g. "secret not found for env://UIFORGE_OMNIROADMAP_DSN").

### Server wiring and flag removal

- Remove `--omniroadmap-dsn` flag, `Config.OmniRoadmapDSN`, and the hard-coded provider block in `newServerInternal`.
- Construct resolver → store → `analytics.Service` → `LoadAll` at startup; `Service.Close` on shutdown.
- `NewOmniRoadmapProvider` and `omniroadmap.go` remain; only their construction moves behind the connector registry.

### Builder UI

New `components/data-sources/AnalyticsSourcePanel.tsx` in the Question workspace (`/builder/?mode=questions`):

- List sources with status badge; add/edit form with name, connector dropdown (from `/connectors`), `dsnRef` input with scheme hint; Test and Remove actions.
- On any mutation, re-fetch `/api/v1/analytics/catalog` so the schema browser updates.
- API client additions in `builder/src/api/dashforge.ts`.

## Security Considerations

- Raw DSNs rejected on input (`IsSecretRef` gate) — nothing secret at rest in either store.
- Resolved secrets: in-memory only, function-scoped, excluded from logs and API responses.
- Existing GrokifyQL policy enforcement (save-time and ad-hoc query) is unchanged and applies uniformly to all sources; SpiceDB-mode dataset/field resource IDs use `SourceConfig.ID`, which operators should keep stable.

## Testing Strategy

Unit tests only (per org convention):

- File store CRUD round-trip and concurrent access.
- Ent store CRUD against an in-memory SQLite ent client.
- Connector registry registration/lookup; unknown-connector error.
- Resolver: `env://` happy path, missing var, raw-DSN rejection.
- Service: add/remove/list with a fake connector factory; catalog merging and ID re-tagging for two sources of the same connector; query routing by source ID.
- API handler tests with the fake-backed service.
