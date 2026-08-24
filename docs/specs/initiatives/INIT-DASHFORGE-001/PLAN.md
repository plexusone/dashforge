# PLAN — DashForge Rename & UIForge Split

**Initiative:** `INIT-DASHFORGE-001`
**Status:** Draft

## Entry Gate (must be true before Phase 1 starts)

1. INIT-UIFORGE-002's working-tree changes are committed across all four
   repos (uiforge — now this repo, godolt, omniroadmap, visionstudio), pushed, and released
   (godolt v0.3.0 tagged; temporary `replace` directives removed).
2. INIT-UIFORGE-002 transitioned to `delivery_complete` or beyond.
3. Disk-space condition that blocked git object writes is resolved.

Rationale: a mass mechanical rename mixed into substantive uncommitted diffs
makes both unreviewable; and the rename touches the same files the pending
commits do.

## Build Order

### Phase 1 — Rename to DashForge

1. GitHub rename `plexusone/uiforge` → `plexusone/dashforge`; move local
   checkout to `~/go/src/github.com/plexusone/dashforge`; update remotes and
   the VisionStudio registry (repoint/retire old entry, fix new entry's
   path).
2. Module path rewrite (`go.mod` + all imports), binary renames
   (`cmd/dashforge-server`, `cmd/dashforge`), `.gitignore` fix, help text.
3. Frontend and docs rebrand: builder package/branding (UF → DF mark), README,
   mkdocs, examples.
4. Consumers: omniroadmap imports/replace → dashforge (cross-repo
   `RMI-OMNIROADMAP-011`); org docs sweep.
5. **Verification gate**: ecosystem-wide grep sweep (no live
   `plexusone/uiforge` references outside historical docs), remotes checked,
   CI green under the new name.

**Exit criteria:** `dashforge-server serve` works identically; all consumers
build; gate checklist recorded in the RMI's handoff notes.

### Phase 2 — UISpec Boundary & New UIForge Repo

6. Boundary inventory doc: per-package classification (dashforge-core vs
   uiforge-extract) with dependency directions.
7. Create fresh `plexusone/uiforge` (retires the redirect — depends on the
   Phase 1 gate) seeded with component-platform PRD/TRD and standard
   scaffolding.
8. First extraction: UISpec/PageSpec types + schema generation into the new
   uiforge; dashforge consumes as a dependency. Further extraction items are
   logged in the new repo's own roadmap as they are sized.

**Exit criteria:** new uiforge builds and tests standalone; dashforge imports
at least one extracted package; boundary doc merged in both repos.

## Risks and Mitigations

See TRD Risks: redirect reuse trap (gated), uncommitted-work collision
(entry gate), module proxy residue (accepted), historical tracking IDs
(immutable by policy).

## Follow-On

- `INIT-DASHFORGE-002` — DashForge Cloud & workspace sync (per-workspace Dolt
  remotes via godolt; control plane; hosted server). Deliberately sequenced
  after the rename so cloud never launches under a transitional name.
- New uiforge repo's own initiative for the component-platform roadmap once
  the boundary doc lands.
