# Analytics Catalog

The analytics catalog is UIForge's neutral description of queryable data across application and database sources. It lets UIForge build Metabase/Looker/Tableau-style schema browsers, query builders, dashboards, and AI-assisted reports without hard-coding a specific product domain.

## Shape

A catalog contains one or more sources. Each source contains datasets, and each dataset contains fields.

```json
{
  "id": "roadmap-app",
  "name": "Roadmap App",
  "sources": [
    {
      "id": "roadmap-app",
      "name": "Roadmap App",
      "type": "application",
      "datasets": [
        {
          "id": "initiatives",
          "name": "Initiatives",
          "queryName": "initiatives",
          "fields": [
            {
              "id": "workspace_ref",
              "name": "Workspace Ref",
              "queryName": "workspace_ref",
              "type": "string",
              "source": "standard",
              "role": "dimension",
              "selectable": true,
              "filterable": true,
              "sortable": true
            },
            {
              "id": "custom.product",
              "name": "Product",
              "queryName": "custom.product",
              "type": "choice",
              "source": "custom",
              "role": "dimension",
              "selectable": true,
              "filterable": true,
              "sortable": true,
              "coverage": 0.8,
              "sampleValues": ["Platform", "Security"]
            }
          ]
        }
      ]
    }
  ]
}
```

## Field Sources

- `standard`: stable fields in the application's canonical model or database table.
- `custom`: tenant/provider-defined fields discovered from records or field definitions.
- `derived`: computed fields such as `rice_score` or `moscow_rank`.
- `metadata`: provider-specific fields exposed intentionally for analysis.

## Query Engines

The catalog is descriptive, not an execution engine. A source can be backed by GrokifyQL, SQL, Ent, Cube, or another query implementation. Hosted SaaS deployments should validate queries against the catalog and bind parameters through the engine instead of interpolating user input into raw SQL strings.

## Managing Sources

Analytics sources are persisted configuration, not startup flags. Each source
is a `SourceConfig`:

```json
{
  "id": "roadmap-app-local",
  "name": "Roadmap App Local",
  "connector": "roadmap-app",
  "dsnRef": "env://UIFORGE_ROADMAP_APP_DSN",
  "enabled": true
}
```

Sources persist in the UIForge metadata database (ent `AnalyticsSource` table)
when `--db-url` is configured, and otherwise in
`.uiforge/analytics-sources.json` (override with `--analytics-source-store`).
Both stores hold only the `dsnRef` secret reference — validation rejects raw
DSNs, so credentials can never reach disk or the metadata database.

Sources are managed at runtime through the API:

| Method | Path | Behavior |
|--------|------|----------|
| GET | `/api/v1/analytics/sources` | List configs with runtime status (`connected`/`error`/`disabled`) |
| POST | `/api/v1/analytics/sources` | Create and connect |
| PUT | `/api/v1/analytics/sources/{id}` | Update and reconnect (ID is immutable) |
| DELETE | `/api/v1/analytics/sources/{id}` | Disconnect and delete |
| POST | `/api/v1/analytics/sources/test` | Probe an unsaved candidate config |
| GET | `/api/v1/analytics/connectors` | List registered connector types |

or through the **Data sources** panel in the Question builder header
(`/builder/?mode=questions`). Enabled sources reconnect automatically at
server startup; a source that fails to connect surfaces as status `error`
without preventing startup.

Connectors are registered in the public `analytics` package via
`RegisterConnector`. The engine core ships no connectors — consuming
application binaries register their own (e.g. a roadmap product registering
its store as a source) and additional connectors register the same way.

## Source Credentials

`dsnRef` is an OmniVault secret reference resolved in memory at connect time.
The resolved DSN is never persisted, logged, returned by the API, or sent to
the browser, and dial errors are sanitized before reaching clients.

Implemented schemes:

- `env://VAR_NAME` — environment variable (development default).
- `file:///absolute/path` — secret file, rooted at the filesystem root.

Planned schemes (register on the same resolver without redesign):

- **Local desktop**: OS keychain via an OmniVault keyring provider module.
- **Server/cloud**: cloud secret managers such as `aws-sm://...`,
  `gcp-sm://...`, or `azure-kv://...` through provider modules.
- **Security-gated deployments**: VaultGuard to enforce posture and provider
  selection before resolving credentials.
- **Self-contained fallback**: OmniVault's SQL store provider with a
  UIForge-owned table such as `uiforge_vault_secrets`, and references like
  `sql://analytics-sources/roadmap-app-local`. The SQL store encrypts secret
  payloads before writing rows; keep the encryption key outside that database.

## Questions

A Question is a saved, read-only analytics query plus display metadata. UIForge
uses the term Question for human-authored analytics artifacts, similar to
Metabase. The query text is still a GrokifyQL query, but Question makes the UI
clear that the artifact is for reporting and cannot mutate data.

The Question builder defaults new Questions into SQL edit mode. Saved Questions
open in a formatted, syntax-highlighted display mode and expose an explicit edit
action. This keeps review and editing separate without showing duplicate SQL
panes.

When a Question is saved, UIForge compiles the GrokifyQL text into metadata:

- parsed AST
- fingerprint
- referenced dataset
- referenced fields
- read-only flag
- requested limit, if provided

Dashboards can reference a Question and choose their own visualization for that
dashboard widget. The same Question can be used in zero, one, or more
dashboards.

Question result tables are treated as analytical output, not page layout. They
scroll horizontally inside the result panel so long text values and wide column
sets remain inspectable without widening the dashboard builder shell.

## Field Values

Filterable catalog fields can expose distinct values for query construction.
UIForge supports two scopes:

- **All values**: distinct values across the selected dataset.
- **Current filter**: distinct values under the current query's `WHERE` clause.

The resulting values can be inserted into the GrokifyQL query as an `IN (...)`
predicate. The same catalog and authorization checks used for query execution
should be applied to field value lookup.

## Authorization

The catalog is also the schema input for analytics authorization. UIForge derives
GrokifyQL entities from datasets and GrokifyQL fields from catalog fields. In
SpiceDB mode, those entities and fields are checked through SystemForge before a
Question is saved or an ad-hoc GrokifyQL query is executed.

See [Analytics Authorization](analytics-authorization.md) for the SpiceDB
resource model and permission mapping.

## Generic Endpoint

Server mode should expose the combined catalog for all connected sources at:

```http
GET /api/v1/analytics/catalog
```

A UIForge server can combine multiple sources — application connectors registered by the consuming binary alongside PostgreSQL and Dolt databases — into one catalog response.
