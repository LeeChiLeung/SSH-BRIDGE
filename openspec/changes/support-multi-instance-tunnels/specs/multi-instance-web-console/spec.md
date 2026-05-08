## ADDED Requirements

### Requirement: Web console cards show profile-specific runtime state
The system SHALL render runtime state, endpoint, health, and errors on each profile card using that profile's own runtime snapshot instead of one global snapshot.

#### Scenario: Two active cards show different states
- **WHEN** two profiles are active and one is `connected` while the other is `retrying`
- **THEN** each card displays its own current state without overwriting the other card's state

### Requirement: Web console start and stop actions target only the selected profile
The system SHALL make per-card start and stop actions operate only on the selected profile.

#### Scenario: Stop one active card
- **WHEN** the operator clicks stop on one active profile card
- **THEN** only that profile stops and other active cards remain running

### Requirement: Web console diagnostics are profile-specific
The system SHALL render the detail panel and diagnostic endpoints for the selected profile using only that profile's endpoint, health, last error, and logs.

#### Scenario: Select diagnostics for a retrying profile
- **WHEN** the operator opens the detail view for a profile that is retrying
- **THEN** the detail panel shows only that profile's retry-related logs and health state

### Requirement: Web console handles concurrent active profiles cleanly
The system SHALL support multiple simultaneously active cards without requiring the operator to deactivate one profile before starting another.

#### Scenario: Start a second profile from the card list
- **WHEN** the operator starts a second profile while one profile is already connected
- **THEN** the second card can progress through `connecting` to `connected` independently and the first card remains active
