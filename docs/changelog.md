# Changelog

All notable changes to UIForge are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html),
and commits follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## [v0.5.0] - 2026-08-17

Static viewer polish: light/dark theme toggle and consistent numeric formatting. Fixed a pre-existing broken build in builder/, added ESLint/Prettier tooling to builder/ and renderer/, and resolved every pre-existing discarded-error finding in datasource/ and internal/server/.

[:octicons-tag-24: Release Notes](releases/v0.5.0.md){ .md-button }

### Highlights

- Static viewer: light/dark theme toggle, persisted via localStorage
- Static viewer: numeric display capped to 2 decimal places, plus a new "currency" metric format
- Resolved all 10 pre-existing "error discard without comment" findings in datasource/ and internal/server/
- Added ESLint and Prettier to builder/ and renderer/, with all resulting findings fixed
- Fixed a pre-existing broken build in builder/: missing @grokify/echartify dependency

### Added

- Light/dark theme toggle in the static viewer header
- "currency" metric format and 2-decimal-place numeric formatting for chart tooltips and axis labels

### Fixed

- Missing `@grokify/echartify` dependency in builder/package.json
- builder/ and renderer/ ESLint findings (hooks, Fast Refresh, dead code, an unwired prop)
- Discarded errors in datasource/ and internal/server/, resolved via proper return/panic handling instead of silent discard

---

## [v0.4.0] - 2026-07-25

UISpec type system, component registry, React renderer, AI-powered builder features, new CLIs, JSON schemas, and domain packages. Completes the DashForge → UIForge rename.

[:octicons-tag-24: Release Notes](releases/v0.4.0.md){ .md-button }

### Highlights

- UISpec type system: page, component, layout, binding, interaction, navigation, theme, capability, and experience profiles
- Component registry with builtin core, analytics, and assistant namespaces
- React PageRenderer for UISpec pages
- AI-powered PageSpec generator and visual builder features
- New `uiforge` and `uiforge-server` Cobra CLIs
- Generated component and page JSON schemas

### Added

- UISpec type system (page, component, layout, binding, interaction, navigation, theme, capability, profiles)
- Component registry with manifest validation and experience profiles
- Domain packages for diff, expression, interaction, state, and bridge
- `uiforge` and `uiforge-server` Cobra CLIs
- Component and page JSON schemas generated via `invopop/jsonschema`
- React PageRenderer with UISpec rendering
- Builder: PageSpec generator, ComponentPalette, ComponentPropertyEditor

### Changed

- Renamed Go module from `dashforge` to `uiforge`, updating import paths across the codebase

---

## [v0.3.0] - 2026-05-11

Migrate from CoreForge to SystemForge with security lint fixes.

[:octicons-tag-24: Release Notes](releases/v0.3.0.md){ .md-button }

### Highlights

- Migrate from CoreForge to SystemForge (upstream project rename)
- Update to SystemForge v0.7.0

### Breaking

- Dependency renamed from `coreforge` to `systemforge` - users of `multiapp` package must update imports

### Fixed

- Resolved gosec G124 and G710 warnings in OAuth handler

---

## [v0.2.0] - 2026-04-26

Principal-based identity model, dashboard marketplace, and multi-app deployment support.

[:octicons-tag-24: Release Notes](releases/v0.2.0.md){ .md-button }

### Highlights

- Principal-based identity model with SystemForge integration
- Dashboard template marketplace with licensing and subscriptions
- Multi-app deployment support via AppBackend adapter

### Added

- Principal entity as unified identity root for all actor types
- Marketplace entities: Publisher, Listing, License, Subscription, SeatAssignment
- DashboardTemplate entity for reusable templates
- SpiceDB integration for fine-grained access control
- Multi-app backend adapter for SystemForge deployment

### Changed

- Migrated JWT and OAuth to SystemForge identity packages
- Replaced custom ChartIR types with @grokify/echartify

---

## [v0.1.0] - 2026-03-01

Initial release with full-stack dashboard builder.

[:octicons-tag-24: Release Notes](releases/v0.1.0.md){ .md-button }

### Highlights

- Full-stack dashboard builder with Go backend and TypeScript frontend
- Ent-based database schema with OAuth authentication
- Data source integrations and alert system

### Added

- Go server with Chi router, JWT auth, and OAuth (GitHub, Google)
- Database entities: User, Organization, Dashboard, SavedQuery, Alert
- DataSource and Integration entities for external connections
- TypeScript ChartIR types and dashboard definitions

---

[v0.4.0]: https://github.com/plexusone/uiforge/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/plexusone/uiforge/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/plexusone/uiforge/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/plexusone/uiforge/releases/tag/v0.1.0
