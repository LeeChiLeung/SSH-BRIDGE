## ADDED Requirements

### Requirement: Web console cards show profile traffic metrics as text
The system SHALL render each tunnel card with text-based traffic metrics for the selected profile snapshot, including current upload speed, current download speed, cumulative upload volume, cumulative download volume, and active connection count.

#### Scenario: Show metrics for an active profile card
- **WHEN** a tunnel profile is connected and carrying traffic
- **THEN** its card displays text values for current upload speed, current download speed, cumulative transferred bytes, and active connection count

#### Scenario: Show idle metrics for a disconnected profile card
- **WHEN** a tunnel profile is disconnected or idle
- **THEN** its card displays zero or empty-equivalent live speed values instead of stale non-zero transfer rates

### Requirement: Web console detail panel shows profile-specific traffic metrics
The system SHALL render the detail or diagnostics view for a selected profile using only that profile's traffic counters and current speed values.

#### Scenario: Open diagnostics for one active profile
- **WHEN** the operator opens the detail view for one profile while another profile is also active
- **THEN** the detail panel shows traffic metrics only for the selected profile

### Requirement: Traffic metrics are available through Web status APIs
The system SHALL include traffic metric fields in the profile snapshot payloads returned by the local Web status APIs.

#### Scenario: Query status for concurrent profiles
- **WHEN** a client requests Web status data while multiple profiles are active
- **THEN** each returned profile snapshot includes its own current upload speed, current download speed, cumulative upload bytes, cumulative download bytes, and active connection count
