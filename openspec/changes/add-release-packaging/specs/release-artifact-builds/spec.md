## ADDED Requirements

### Requirement: The repository provides repeatable multi-platform binary builds
The system SHALL provide a repository-owned build entrypoint that produces release binaries for a defined target matrix without requiring the operator to compose raw `go build` commands manually.

#### Scenario: Build all configured release binaries
- **WHEN** the operator runs the release build entrypoint
- **THEN** the repository produces binaries for each configured platform and architecture target in the release matrix

### Requirement: Release outputs use a predictable artifact directory layout
The system SHALL write generated release artifacts to a standardized `dist/` directory using deterministic names that encode platform and architecture.

#### Scenario: Inspect built artifacts after a release build
- **WHEN** the release build entrypoint completes successfully
- **THEN** the generated binaries appear under `dist/` with stable names that identify their target platform and architecture

### Requirement: Release build scripts support versioned artifact naming
The system SHALL allow the release build workflow to name artifacts with a version string that can be supplied by the operator or defaulted for development builds.

#### Scenario: Build with an explicit version
- **WHEN** the operator invokes the release workflow with a version value
- **THEN** the generated artifact names include that version value consistently across the produced binaries
