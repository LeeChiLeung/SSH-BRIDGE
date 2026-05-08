## MODIFIED Requirements

### Requirement: Web console can configure tunnel parameters
The system SHALL provide browser-based profile creation and editing controls for SSH address, SSH port, SSH user, SSH password, local bind address, local port, remote target address, and private key path for the tunnel.

#### Scenario: Create a profile from the add action
- **WHEN** the operator clicks the `新增` action
- **THEN** the console opens a creation flow with editable tunnel configuration fields for the supported SSH and forwarding parameters, including SSH password and private key path

#### Scenario: Edit tunnel configuration for an existing profile
- **WHEN** the operator triggers the edit action for an existing profile
- **THEN** the console loads that profile's current tunnel settings into the editing flow, including any saved SSH password value and private key path

### Requirement: Web console can save and reuse tunnel profiles
The system SHALL allow operators to save named tunnel profiles locally and reload them in later sessions, including the configured SSH password and private key path values used for authentication.

#### Scenario: Save a profile
- **WHEN** the operator saves the current configuration under a profile name
- **THEN** the profile is persisted locally with its configured tunnel settings, including SSH authentication fields, and appears in the saved profile list

#### Scenario: Load a saved profile
- **WHEN** the operator selects a previously saved profile
- **THEN** the console repopulates the configuration form with that profile's saved tunnel settings, including SSH password and private key path values

#### Scenario: Show saved profiles on the list home page
- **WHEN** the operator returns to the home page in a later session
- **THEN** the saved profiles appear again as cards in the list workspace
