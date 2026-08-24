# PRD — DashForge Rename & UIForge Split — Analytics Platform Reclaims Its Name

**Initiative:** `INIT-DASHFORGE-001`
**Status:** Draft (blocked on landing INIT-UIFORGE-002)
**Home repo:** `github.com/plexusone/dashforge` (formerly `github.com/plexusone/uiforge`)

## Problem

The repository has drifted from its name twice:

1. It began as **dashforge** — "a JSON-first dashboard framework that … grows
   into a full Metabase-like analytics platform" (still the live GitHub repo
   description).
2. It was renamed to **uiforge** when the work centered on UI widgets, with a
   Salesforce-Lightning-style component-platform vision (UISpec, PageSpec,
   component namespaces — see `docs/specs/{PRD,TRD}.md`).
3. INIT-UIFORGE-002 then delivered the original vision anyway: a working
   analytics platform (multi-source catalog, Questions, dashboards, persistent
   sources with OmniVault, Dolt metadata DB). The internals never fully left
   the old name — the server help text says "Dashforge server", the API client
   is `dashforge.ts`, the db package doc says "Dashforge Server".

Result: the analytics platform ships under a name that describes a different
product, and the component-platform vision has no home of its own.

## Goals

- **Rename the repo back to `dashforge`** — GitHub rename (reversing the
  earlier rename; the `dashforge` redirect alias already points here), Go
  module path, binaries, frontend branding, docs, and every consumer import.
- **Create a new `uiforge` repo** for the UI-customization platform: UISpec
  type system, PageSpec, renderer, component namespaces — seeded from the
  original UIForge specs, populated by incremental extraction (not a bulk
  move of entangled code).
- **Break nothing in the ecosystem**: omniroadmap (which imports this module),
  VisionStudio registry/RMI tracking, org docs, CI, and local clones all
  updated in a verified order.

## Non-Goals

- DashForge Cloud / workspace sync — follows as `INIT-DASHFORGE-002` after
  this initiative, precisely so cloud never launches under a name about to
  change.
- Completing the extraction of all rendering code into the new uiforge — this
  initiative establishes the boundary and moves the first clean units; deeper
  extraction is follow-on work in the new repo's own roadmap.
- Renaming historical tracking artifacts: `INIT-UIFORGE-001/-002` and
  `RMI-UIFORGE-*` IDs remain as historical record.

## Users and Experience

- **Operators/analysts**: `dashforge-server serve …` with identical behavior;
  README and docs describe an analytics platform named DashForge.
- **Ecosystem developers**: `import "github.com/plexusone/dashforge/..."`;
  the old module path disappears in one coordinated sweep, not gradually.
- **UI-platform developers**: a fresh `plexusone/uiforge` whose README/specs
  describe the component platform, without analytics code tangled in.

## Sequencing Constraint (one-way door)

GitHub's rename redirect (`uiforge` → `dashforge`) dies the moment a new repo
named `uiforge` is created. Any stale remote, CI reference, or `go get` would
then silently resolve to the **new, different** repo. Therefore: rename →
update every reference → verify nothing still resolves through the redirect →
only then create the new dashforge repo. This ordering is enforced as RMI
dependencies, not convention.

## Success Criteria

- `github.com/plexusone/dashforge` is the canonical repo; module path, binary
  names (`dashforge-server`, `dashforge`), and builder/docs branding match.
- omniroadmap and all local consumers build against the new path with no
  `uiforge` references to the analytics code remaining (verified by sweep).
- New `plexusone/uiforge` exists with the component-platform specs and the
  first extracted UISpec units; dashforge consumes it as a dependency.
- VisionStudio registry reflects the rename; new work logs as
  `RMI-DASHFORGE-*`.
