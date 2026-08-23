# CLAUDE.md — uiforge

UIForge is a specification-driven UI composition platform. JSON specifications (UISpec) are the primary artifact — Go types are the source of truth, React renders them.

## Specs

Design decisions are documented in `docs/specs/` — read before implementing:

- `docs/specs/PRD.md` — problem, goals, non-goals, experience profiles
- `docs/specs/TRD.md` — architecture, UISpec type system, component namespaces, repository structure
- `docs/specs/PLAN.md` — phased build order with exit criteria and risks
- `docs/specs/ROADMAP.md` — RMI-level breakdown by themed phase

## Stack

- **Go module:** `github.com/plexusone/uiforge` (renaming from `dashforge`)
- **ORM:** Ent (`entgo.io/ent`) for any persistence needs
- **CLI:** Cobra (`github.com/spf13/cobra`)
- **Frontend:** React + TypeScript (in `ts/` and `builder/`)
- **JSON Schema:** generated from Go types via `invopop/jsonschema`, linted with `schemago`
- **Query safety:** GrokifyQL for saved Questions and analytics queries
- **Authorization:** SystemForge `authz.Authorizer`, with SpiceDB mode for fine-grained permissions

## Analytics Notes

- Saved Questions are read-only GrokifyQL artifacts persisted by the UIForge
  backend. Dashboard widgets reference Questions and choose their own
  visualization.
- The Question builder intentionally separates SQL display and editing:
  existing saved Questions open in formatted, syntax-highlighted read mode;
  new Questions and query mutations switch to edit mode. Result tables should
  scroll horizontally inside the middle column for wide result sets.
- The analytics catalog is the source of truth for queryable sources, datasets,
  and fields. Use `dashboardir.AnalyticsCatalog` and
  the public `analytics` package instead of hard-coding application-specific
  assumptions into UI code.
- Persist analytics source connection metadata in UIForge's metadata DB, but
  store only secret references (`env://...`, `keyring://...`, `aws-sm://...`,
  etc.). Resolve them at runtime with OmniVault; use VaultGuard when posture and
  provider-selection policy are needed. Never return resolved DSNs/tokens to the
  browser.
- If a self-contained encrypted-in-DB fallback is needed, use OmniVault's SQL
  store provider (`sql://...`) with a UIForge-owned table and an encryption key
  sourced outside that database. Do not create a parallel UIForge-only secret
  interface.
- Agent marketplace concepts live in `github.com/plexusone/omniagent`, not in
  UIForge. UIForge should consume OmniAgent marketplace providers and publish
  app-owned capability names such as `uiforge.query.run`,
  `uiforge.question.write`, and `uiforge.field_values.read`.
- GrokifyQL policy is enforced both when saving Questions and when executing
  ad-hoc `/api/v1/analytics/query` requests.
- In SpiceDB mode, UIForge uses SystemForge as the bridge and maps catalog
  datasets/fields to `analytics_dataset` and `analytics_field` resources.
  Keep resource ID generation stable when changing query names.

## Known Gotchas

- **`cube/` has no lint/typecheck/format tooling, by design.** It's a thin
  Cube.js semantic-layer service directory (`@cubejs-backend/*` deps only,
  `build` script is a no-op echo) — it was never meant to be a
  typechecked/linted TS app like `builder/` or `renderer/`. `atrelease
  release`'s validation auto-detects it as a TS project anyway (it has
  `.ts` files) and fails on missing eslint config / no `tsc`. This is a
  false positive, not real debt — don't add tooling to `cube/` to silence
  it; run `atrelease release <version> --skip-checks` instead (after
  manually verifying `go build`/`go test`/`golangci-lint`/`builder/` and
  `renderer/`'s own lint+format are actually clean).
- **CI can't run a newer Go patch than what it cached.** The shared
  `plexusone/.github` reusable `go-ci.yaml` (`actions/setup-go@v7`)
  resolves `go-version: "1.26.x"` to whatever patch it last cached and
  pins `GOTOOLCHAIN=local` — it does not auto-upgrade just because
  `go.mod`'s `go` directive asks for a newer patch. Bumping `go.mod` past
  CI's cached version breaks the build with `go: go.mod requires go >=
  X (running go Y; GOTOOLCHAIN=local)`. Check CI's actual resolved
  version before bumping (or fix the shared workflow to
  `check-latest: true`) — don't just bump-and-push. See the same note in
  the `plexusone` org CLAUDE.md.

## PRISM Control

This repo's roadmap items are tracked in [prism-control](https://github.com/ProductBuildersHQ/prism-control). The initiative is **INIT-UIFORGE-001** with 37 RMIs across 6 phases.

### Session Protocol

Every Claude Code session that implements work follows this check-in/check-out protocol:

#### 1. Find work

```bash
prismctl work ready --initiative INIT-UIFORGE-001
```

#### 2. Claim (check out)

Claim an RMI before starting. This creates a lease-based assignment:

```bash
prismctl work claim RMI-UIFORGE-001 --worker "session-$(date +%s)" --lease-hours 4
```

Transition the RMI to in-progress:

```bash
prismctl rmi update RMI-UIFORGE-001 --status in_progress
```

Renew the lease if work takes longer:

```bash
prismctl work renew <assignment-id> --lease-hours 4
```

#### 3. Execute

Work in this repo. Every commit carries the RMI trailer:

```text
feat(uispec): define core PageSpec types

Refs: RMI-UIFORGE-001
```

Trailer rules:

- Git trailer with key `Refs:` — value is one RMI ID
- RMI only, never the initiative ID
- Subject line stays clean; trailer goes in the footer
- Conventional Commits format for the message itself

#### 4. Complete (check in)

```bash
# Add evidence and handoff notes
prismctl work update <assignment-id> \
  --handoff '{"completed":["module rename"],"remaining":[],"decisions":[],"next_action":"none"}'

# Mark complete and auto-transition RMI to completed
prismctl work complete <assignment-id> --transition
```

Or release if handing off to another session:

```bash
prismctl work release <assignment-id> \
  --handoff '{"completed":["core logic"],"remaining":["tests"],"next_action":"add unit tests"}'
```

### Cross-Repo RMIs

Some RMIs land in sibling repos (`plexusone/agentos`, `plexusone/agentos-web`). The RMI ID reflects where the code lands, not the initiative home:

- `RMI-UIFORGE-*` — work in this repo
- `RMI-AGENTOS-*` — work in agentos or agentos-web
