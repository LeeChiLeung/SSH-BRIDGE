## ADDED Requirements

### Requirement: Web console home page is a card-based tunnel list
The system SHALL present the local Web console home page as a list of saved tunnel cards with a title area, a top-level add action, and a main workspace for browsing multiple tunnel profiles.

#### Scenario: Load the list home page
- **WHEN** the operator opens the local Web console
- **THEN** the page displays a title area, a `新增` action, and a list of saved tunnel cards in the main workspace

#### Scenario: Render summary cards for multiple profiles
- **WHEN** the operator has more than one saved tunnel profile
- **THEN** the home page displays one card per saved profile instead of collapsing all profiles into a single shared form

### Requirement: Each tunnel card shows operational summary fields
The system SHALL display summary information on each tunnel card including SSH address, local proxy port, remote port, and last enable time.

#### Scenario: Show summary information on a card
- **WHEN** a tunnel profile appears in the list
- **THEN** its card displays the SSH address, local proxy port, remote port, and last enable time in a dedicated information area

#### Scenario: Show an unset last enable time
- **WHEN** a profile has never been started successfully
- **THEN** the card indicates that no enable time is available yet

### Requirement: Web console can configure tunnel parameters
The system SHALL provide browser-based profile creation and editing controls for SSH address, SSH port, SSH user, local bind address, local port, remote target address, and private key path for the tunnel.

#### Scenario: Create a profile from the add action
- **WHEN** the operator clicks the `新增` action
- **THEN** the console opens a creation flow with editable tunnel configuration fields for the supported SSH and forwarding parameters

#### Scenario: Edit tunnel configuration for an existing profile
- **WHEN** the operator triggers the edit action for an existing profile
- **THEN** the console loads that profile's current tunnel settings into the editing flow

### Requirement: Web console can start and stop the tunnel
The system SHALL allow the operator to start and stop the managed tunnel instance from the local Web console without requiring manual CLI invocation.

#### Scenario: Start a tunnel from the console
- **WHEN** the operator submits a valid tunnel configuration and triggers the start action
- **THEN** the system starts the tunnel runtime and updates the console with the new runtime state

#### Scenario: Stop a tunnel from the console
- **WHEN** the operator triggers the stop action while a tunnel is active or retrying
- **THEN** the system shuts down the managed tunnel runtime and updates the console to a stopped state

### Requirement: Web console can duplicate profiles
The system SHALL allow the operator to duplicate an existing tunnel profile from the list so similar configurations can be reused quickly.

#### Scenario: Duplicate an existing profile
- **WHEN** the operator triggers the duplicate action on a tunnel card
- **THEN** the system creates a new profile initialized from the original profile's saved tunnel settings

### Requirement: Web console shows live runtime state
The system SHALL display the current managed tunnel state and SHALL surface at least `disconnected`, `connecting`, `connected`, and `retrying`.

#### Scenario: Show a connecting state
- **WHEN** the tunnel runtime is attempting to establish an SSH session
- **THEN** the console displays the `connecting` state

#### Scenario: Show a retrying state
- **WHEN** the tunnel runtime encounters a retryable failure and is waiting for the next reconnect attempt
- **THEN** the console displays the `retrying` state

#### Scenario: Show a connected state
- **WHEN** the SSH control connection is established and ready to forward traffic
- **THEN** the console displays the `connected` state

### Requirement: Web console shows operational logs in a detail area
The system SHALL provide a log and detail area associated with a selected or expanded tunnel that displays recent runtime events including reconnect attempts, keepalive failures, authentication failures, and network timeouts.

#### Scenario: Display reconnect log entries
- **WHEN** the operator opens or selects the detail view for a tunnel whose runtime enters reconnect logic after a retryable failure
- **THEN** the detail area includes a recent entry describing the failure and retry timing

#### Scenario: Display authentication failure
- **WHEN** the operator opens or selects the detail view for a tunnel whose runtime fails due to invalid credentials or authentication rejection
- **THEN** the detail area includes an authentication-related error entry

### Requirement: Web console shows current tunnel endpoint and health
The system SHALL display the current local listener endpoint on the relevant tunnel card and SHALL expose health information for local listener readiness and SSH control-plane connectivity.

#### Scenario: Show the current local entrypoint
- **WHEN** the tunnel listener is bound successfully
- **THEN** the console displays the effective local endpoint, including the bound address and port

#### Scenario: Show unhealthy SSH connectivity
- **WHEN** the local listener remains bound but the SSH control connection is unavailable
- **THEN** the console indicates that the listener and SSH connectivity health are not both healthy

### Requirement: Web console can save and reuse tunnel profiles
The system SHALL allow operators to save named tunnel profiles locally and reload them in later sessions.

#### Scenario: Save a profile
- **WHEN** the operator saves the current configuration under a profile name
- **THEN** the profile is persisted locally and appears in the saved profile list

#### Scenario: Load a saved profile
- **WHEN** the operator selects a previously saved profile
- **THEN** the console repopulates the configuration form with that profile's saved tunnel settings

#### Scenario: Show saved profiles on the list home page
- **WHEN** the operator returns to the home page in a later session
- **THEN** the saved profiles appear again as cards in the list workspace

### Requirement: Web console uses lightweight server-driven UI technology
The system SHALL implement the first-version Web console using Go HTTP handlers, server-rendered templates, and HTMX-style partial updates rather than Vue, React, or other SPA frameworks.

#### Scenario: Render the page without a SPA framework
- **WHEN** the operator opens the local Web console in a browser
- **THEN** the page is served by the Go application as server-rendered HTML and enhanced with lightweight partial-update behavior
