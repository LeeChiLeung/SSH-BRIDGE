## ADDED Requirements

### Requirement: List badges reflect tunnel state with distinct colors
The system SHALL render the tunnel state badge in the new control console list using distinct visual colors for disconnected, connecting, connected, and failed states.

#### Scenario: Disconnected tunnel is gray
- **WHEN** a tunnel is disconnected and has no current error
- **THEN** the list state badge MUST be rendered using a gray visual style

#### Scenario: Connecting tunnel is blue
- **WHEN** a tunnel is connecting or retrying
- **THEN** the list state badge MUST be rendered using a blue visual style

#### Scenario: Connected tunnel is green
- **WHEN** a tunnel is connected
- **THEN** the list state badge MUST be rendered using a green visual style

#### Scenario: Failed tunnel is red
- **WHEN** a tunnel is currently disconnected and has a last error recorded
- **THEN** the list state badge MUST be rendered using a red visual style
