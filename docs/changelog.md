# Changelog

All notable changes to DashForge are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html),
and commits follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

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

[v0.3.0]: https://github.com/plexusone/dashforge/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/plexusone/dashforge/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/plexusone/dashforge/releases/tag/v0.1.0
