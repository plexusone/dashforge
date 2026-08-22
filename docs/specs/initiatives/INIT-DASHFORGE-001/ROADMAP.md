# ROADMAP — DashForge Rename & UIForge Split

**Initiative:** `INIT-DASHFORGE-001`
**Repository:** `github.com/plexusone/dashforge`

Entry gate: INIT-UIFORGE-002 committed, pushed, and released first (see
PLAN.md). `RMI-OMNIROADMAP-011` is a cross-repo RMI created directly in
VisionStudio (roadmap import assigns a single repository; listed here for
reference only).

## Phase 1 — Rename to DashForge

**Theme:** GitHub, module path, binaries, branding, consumers — in verified order

- [ ] `RMI-DASHFORGE-001` GitHub rename, local checkout move, remotes and VisionStudio registry update
- [ ] `RMI-DASHFORGE-002` Go module path rewrite, binary renames (dashforge-server, dashforge), gitignore and help-text fixes
  - Depends on: `RMI-DASHFORGE-001`
- [ ] `RMI-DASHFORGE-003` Frontend and docs rebrand: builder branding, README, mkdocs, examples
  - Depends on: `RMI-DASHFORGE-002`
- [ ] `RMI-OMNIROADMAP-011` (repo `github.com/grokify/omniroadmap`) Switch imports and replace directives from plexusone/uiforge to plexusone/dashforge
- [ ] `RMI-DASHFORGE-004` Verification gate: ecosystem grep sweep, remotes audit, CI green under new name
  - Depends on: `RMI-DASHFORGE-003`

## Phase 2 — UISpec Boundary & New UIForge Repo

**Theme:** Define the component-platform boundary; reuse the freed uiforge name deliberately

- [ ] `RMI-DASHFORGE-005` Boundary inventory: classify renderer, ts, UISpec/PageSpec, builder widgets as core vs extract
- [ ] `RMI-DASHFORGE-006` Create fresh plexusone/uiforge repo seeded with component-platform specs and scaffolding
  - Depends on: `RMI-DASHFORGE-004`
  - Depends on: `RMI-DASHFORGE-005`
- [ ] `RMI-DASHFORGE-007` First extraction: UISpec/PageSpec types and schema generation into new uiforge; dashforge consumes
  - Depends on: `RMI-DASHFORGE-006`
