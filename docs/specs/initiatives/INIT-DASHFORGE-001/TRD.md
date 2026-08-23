# TRD — DashForge Rename & DashForge Split

**Initiative:** `INIT-DASHFORGE-001`
**Status:** Draft (blocked on landing INIT-UIFORGE-002)

## Current State

- Repo `github.com/plexusone/dashforge`; GitHub redirect alias
  `github.com/plexusone/dashforge` → this repo (verified: `gh repo view
  plexusone/dashforge` returns `"name": "dashforge"`). The redirect is the
  artifact of the original dashforge→dashforge rename and is load-bearing for
  the plan: renaming back reverses it, and the freed `dashforge` name can be
  reused — but only after the redirect is provably unused.
- Go module `github.com/plexusone/dashforge`, v0 (no major-version path suffix
  to manage; the path change is still breaking for consumers).
- Known consumers importing the module: `grokify/omniroadmap` (via local
  `replace`). VisionStudio registers the repo; org CLAUDE.md files and
  ecosystem docs reference it by name.
- Residual dashforge naming already in-tree (helps the rename, must be made
  consistent rather than removed): server help text, `internal/server/db`
  package docs, `builder/src/api/dashforge.ts`, `DashforgeApiError`,
  `docs/specs/PRD.md` history.
- The component-platform vision (UISpec, PageSpec, component namespaces,
  `renderer/`, `ts/`) lives in `docs/specs/{PRD,TRD}.md` and partially in
  code, entangled with dashboard rendering.

## Design

### Rename mechanics (Phase 1)

1. **GitHub**: `gh repo rename dashforge` on `plexusone/dashforge`. History,
   issues, stars carry over; `dashforge` becomes the redirect.
2. **Local**: move checkout to `~/go/src/github.com/plexusone/dashforge`,
   update `origin`; update VisionStudio registry path for
   `github.com/plexusone/dashforge` and retire the old registry entry.
3. **Module path**: `go.mod` module → `github.com/plexusone/dashforge`; every
   in-repo import rewritten (mechanical `gofmt -r` / sed sweep + build).
4. **Binaries**: `cmd/dashforge-server` → `cmd/dashforge-server`,
   `cmd/dashforge` → `cmd/dashforge`. Cobra `Use`/help text, `.gitignore`
   entries (fixing the over-broad `dashforge-server` pattern that currently
   matches `cmd/dashforge-server/`), embed paths.
5. **Frontend/docs**: builder package name, UF logo mark → DF, UI strings,
   README, mkdocs, `examples/`. `dashforge.ts` and `DashforgeApiError` are
   already correctly named.
6. **Consumers**: omniroadmap `go.mod` require/replace + imports switch to
   `dashforge` (cross-repo RMI). Org-level docs sweep
   (`plexusone/.github/CLAUDE.md`, ecosystem references).

### Verification gate (Phase 1 exit)

Before the `dashforge` name may be reused:

- `grep -r "plexusone/dashforge"` across `~/go/src/github.com` returns only
  historical documents (specs, changelogs) — no `go.mod`, import, CI, remote,
  or registry references.
- `git remote -v` in the renamed checkout and all consumers points at
  `dashforge`.
- CI green on the renamed repo (the shared `go-ci.yaml` workflow re-run under
  the new name).

### DashForge split (Phase 2)

1. **Boundary inventory** (doc, in dashforge): classify each candidate —
   `renderer/`, `ts/`, UISpec/PageSpec types, builder widget components —
   as *dashforge-core* (analytics-specific rendering) or *dashforge-extract*
   (generic UI composition). Decision recorded per package with its
   dependency direction; the invariant is dashforge → dashforge, never the
   reverse.
2. **New repo**: create `plexusone/dashforge` fresh (this deliberately retires
   the redirect — hence the Phase 1 gate). Seed with the component-platform
   PRD/TRD (adapted from the original DashForge specs), module
   `github.com/plexusone/dashforge`, standard plexusone scaffolding (CI, lint,
   CLAUDE.md).
3. **First extraction**: move the cleanest UISpec units (types + schema
   generation first, renderer pieces as they untangle) into dashforge;
   dashforge consumes them as a normal dependency. Incremental by design — a
   bulk move would relocate the entanglement rather than resolve it.

### Tracking migration

- VisionStudio: `github.com/plexusone/dashforge` registered (done at
  initiative creation); after the rename, the old
  `github.com/plexusone/dashforge` registry entry is repointed/archived, and —
  once the new dashforge repo exists — re-registered at the new local path for
  the component platform. New RMIs use `RMI-DASHFORGE-*` (and later
  `RMI-DASHFORGE-*` numbering continues from 054 for the new repo's work).
- Historical `INIT-DASHFORGE-*` / `RMI-DASHFORGE-*` records are immutable.

## Risks

- **Redirect reuse trap**: mitigated by the Phase 1 verification gate as a
  hard RMI dependency (see PRD "Sequencing Constraint").
- **Uncommitted-work collision**: the rename must not begin until
  INIT-UIFORGE-002's working-tree changes are committed and released;
  mixing a mass mechanical rename into substantive diffs makes both
  unreviewable. Enforced as the initiative's entry gate in PLAN.md.
- **Go module proxy cache**: `proxy.golang.org` retains
  `github.com/plexusone/dashforge` versions; old tags remain fetchable under
  the old path. Acceptable — consumers are in-house and switch in Phase 1;
  no new tags are ever cut under the old path.
- **prism-control / PRISM tracking**: INIT-UIFORGE-001 lives in
  prism-control's records with the old repo name; treated as historical, not
  rewritten.

## Testing Strategy

- Rename phases: `go build ./... && go test ./...` in dashforge and every
  consumer after each mechanical step; builder `typecheck`/`lint`/`build`;
  the Phase 1 grep-sweep gate.
- Split phase: new dashforge repo carries its own unit tests from the first
  extracted package; dashforge CI proves consumption of the extracted units.
