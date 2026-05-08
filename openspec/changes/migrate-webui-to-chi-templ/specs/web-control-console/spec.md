## MODIFIED Requirements

### Requirement: Web console uses lightweight server-driven UI technology
The system SHALL implement the Web console as a server-rendered local UI without introducing Vue, React, or other SPA frameworks. During the migration away from the monolithic legacy implementation, the system SHALL support both the existing Go-template-based console and a new `chi + templ` console path until the new UI reaches feature parity.

#### Scenario: Render the page without a SPA framework
- **WHEN** the operator opens the local Web console in a browser
- **THEN** the page is served by the Go application as server-rendered HTML and enhanced with lightweight partial-update behavior

#### Scenario: Run with a legacy console during migration
- **WHEN** the maintainer or operator uses the legacy console path while the new UI is still under construction
- **THEN** the application continues to serve the existing `web.go` implementation without requiring the new UI to be fully complete

#### Scenario: Run with the new console path during migration
- **WHEN** the maintainer or operator selects the new console path
- **THEN** the application serves the `chi + templ` implementation while still using the shared tunnel-management core
