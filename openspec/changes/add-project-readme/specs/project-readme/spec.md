## ADDED Requirements

### Requirement: Repository README describes current product capabilities
The repository SHALL include a root-level README that explains the current purpose of the project, its primary tunnel-management capabilities, and the supported operator experience exposed by the codebase.

#### Scenario: Reader opens the repository home page
- **WHEN** a reader views the repository without prior context
- **THEN** the README describes that the project is a Go-based SSH tunneling tool with local forwarding, multi-instance profiles, a local Web console, and a system tray entry

### Requirement: Repository README provides executable startup guidance
The README SHALL document the main ways to run the application, including direct tunnel mode, Web console mode, tray mode, and version inspection, using commands that match the current CLI.

#### Scenario: Reader wants to start the application locally
- **WHEN** the reader follows the README quick-start and usage sections
- **THEN** the README provides concrete commands that align with the current flags and defaults implemented by the binary

### Requirement: Repository README documents packaging and version behavior
The README SHALL explain how release artifacts are built, which targets are included in the default build matrix, what package formats are produced, and how the version string is derived or overridden.

#### Scenario: Maintainer prepares a release build
- **WHEN** the maintainer consults the README for packaging steps
- **THEN** the README documents `scripts/release.sh`, the default Linux/Windows targets, `.deb` and AppImage outputs, and the use of `VERSION=...` or git-derived versions

### Requirement: Repository README documents runtime persistence and limitations
The README SHALL describe where tunnel profiles are stored, how default runtime values behave, and any current known platform limitations that materially affect operation or packaging.

#### Scenario: Operator wants to understand local state and constraints
- **WHEN** the operator reads the configuration and troubleshooting sections
- **THEN** the README states the profile storage behavior, default Web address, and the current limitation around default macOS packaging support
