## ADDED Requirements

### Requirement: Project provides a new chi+templ UI shell under ui directory
The system SHALL provide a new WebUI implementation under a root-level `ui/` directory that uses `chi` for routing and `templ` for server-rendered UI composition.

#### Scenario: Initialize the new UI shell
- **WHEN** a maintainer begins migrating the Web console away from the monolithic legacy implementation
- **THEN** the repository contains a dedicated `ui/` structure for routes, pages, components, and supporting assets built around `chi` and `templ`

### Requirement: New UI shell reuses existing tunnel management core
The new UI shell SHALL reuse the existing tunnel controller, profile store, and runtime state model instead of introducing a second, duplicated business logic implementation.

#### Scenario: Render data in the new UI
- **WHEN** the new UI displays profile lists, tunnel status, or runtime diagnostics
- **THEN** it reads those values from the same shared core modules already used by the legacy UI

### Requirement: New UI shell supports phased page migration
The new UI shell SHALL allow individual Web console pages or panels to be migrated incrementally without requiring the legacy UI to be removed in the same change.

#### Scenario: Migrate one page at a time
- **WHEN** the maintainer ports only a subset of the Web console to the new UI shell
- **THEN** the application can still serve the remaining legacy views until the rest of the migration is completed
