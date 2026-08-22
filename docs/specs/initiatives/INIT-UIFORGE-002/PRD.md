# PRD — Persistent Analytics Sources — Metabase-Style Multi-Source Catalog with OmniVault

**Initiative:** `INIT-UIFORGE-002`
**Status:** Draft
**Home repo:** `github.com/plexusone/uiforge`

## Problem

UIForge's analytics catalog currently supports exactly one application source, and only when the operator passes `--omniroadmap-dsn` on the command line at startup:

```bash
uiforge-server serve --address 127.0.0.1:13319 \
  --omniroadmap-dsn 'root:@tcp(127.0.0.1:13307)/omniroadmap' --enable-ollama
```

This has three problems:

1. **Single source.** UIForge is meant to work like Metabase — one dashboard/question tool connected to many datasources at once. The current wiring hard-codes one OmniRoadmap provider.
2. **Startup-only configuration.** Sources cannot be added, removed, or edited while the server runs. Metabase, Tableau, and Looker all treat datasource management as an in-app admin function.
3. **Credentials on the command line.** Raw DSNs in shell history and process listings are a credential-hygiene problem, and they cannot be persisted safely as-is.

## Goals

- Start the server with no per-source flags: `uiforge-server serve --address 127.0.0.1:13319 --enable-ollama`.
- Analytics sources are persisted configuration: id, display name, connector type, secret reference (`dsnRef`), and enabled flag.
- Sources can be added, edited, tested, and removed at runtime through a CRUD API and a builder UI panel, and survive server restarts.
- Multiple sources are served as one combined catalog to the Question builder (`/builder/?mode=questions`).
- Credentials are never persisted or returned to the browser. Only secret references (`env://…`, `file://…`, later `keyring://…`, `aws-sm://…`, `sql://…`) are stored; they are resolved in memory at connect time via OmniVault v0.5.0.

## Non-Goals

- Raw-DSN entry in the UI backed by OmniVault's `sqlstore`/keyring providers (follow-up initiative; this pass requires the user to supply a secret reference).
- New connector types beyond the existing OmniRoadmap connector (the design must make adding connectors trivial, but VisionStudio/Postgres/etc. connectors are follow-ups).
- Multi-tenant/organization scoping of sources (single-operator model for now, consistent with the rest of the local server).
- Migrating the existing generic `DataSource` ent table / `datasource.Manager` SQL-connection path; analytics sources are a separate concept (application catalogs, not raw SQL connections).

## Users and Experience

**Primary user:** the UIForge operator/analyst running the server locally or on a small team server.

Experience walkthrough:

1. Operator starts the server with no source flags.
2. In the builder's Question workspace, a **Data Sources** panel lists configured sources (empty on first run).
3. Operator clicks **Add source**, picks connector `omniroadmap`, names it, and enters a secret reference such as `env://UIFORGE_OMNIROADMAP_DSN`.
4. **Test connection** resolves the reference in memory, pings the store, and reports success/failure without echoing the DSN.
5. On save, the source appears in the catalog; the Question builder's schema browser now shows its datasets and fields.
6. Restarting the server reconnects all enabled sources automatically.
7. Removing a source drops it from the catalog and closes its connections.

## Success Criteria

- `--omniroadmap-dsn` flag is removed; the documented start command has no source-specific flags.
- A source added via UI/API is present after server restart (both with and without a metadata DB configured).
- Two or more sources appear merged in `GET /api/v1/analytics/catalog` and are independently queryable via `POST /api/v1/analytics/query`.
- No raw DSN ever appears in: persisted stores, API responses, or logs.
- Unit tests cover store round-trips, secret resolution (`env://` scheme), and dynamic add/remove.
