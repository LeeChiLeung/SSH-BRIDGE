## ADDED Requirements

### Requirement: The repository can produce a Debian package
The system SHALL provide a packaging workflow that generates a `.deb` artifact for the application from repository-managed scripts and metadata.

#### Scenario: Build a Debian package
- **WHEN** the operator runs the Debian packaging entrypoint with required packaging tools available
- **THEN** the repository produces a `.deb` package in the standardized release output directory

### Requirement: The repository can produce an AppImage package
The system SHALL provide a packaging workflow that assembles an AppDir and generates an AppImage artifact for Linux distribution.

#### Scenario: Build an AppImage artifact
- **WHEN** the operator runs the AppImage packaging entrypoint with required tooling available
- **THEN** the repository produces an AppImage artifact in the standardized release output directory

### Requirement: Packaging workflows fail clearly when prerequisites are missing
The system SHALL detect missing external packaging tools and SHALL stop with actionable error output rather than silently producing incomplete artifacts.

#### Scenario: Packaging tool is unavailable
- **WHEN** the operator runs Debian or AppImage packaging without the required external packaging tool installed
- **THEN** the workflow exits with an error that identifies the missing tool and the package type that could not be built
