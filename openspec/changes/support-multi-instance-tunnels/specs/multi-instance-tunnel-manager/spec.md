## ADDED Requirements

### Requirement: The application can run multiple tunnel profiles concurrently
The system SHALL allow more than one tunnel profile to be active at the same time within the same application process, provided their local listener bindings do not conflict.

#### Scenario: Start two distinct profiles successfully
- **WHEN** the operator starts two profiles with different local listener ports
- **THEN** both tunnel instances remain active concurrently and each profile keeps its own listener and SSH connection lifecycle

#### Scenario: Starting one profile does not stop another
- **WHEN** one tunnel profile is already active and the operator starts a second profile
- **THEN** the first profile remains active unless the operator explicitly stops it

### Requirement: Runtime state is isolated per profile
The system SHALL track tunnel state, local endpoint, health, last error, and recent logs independently for each running profile.

#### Scenario: One profile retries while another remains connected
- **WHEN** one profile experiences a retryable SSH failure while another profile stays healthy
- **THEN** the failing profile enters `retrying` without changing the connected state of the healthy profile

#### Scenario: One profile stops without affecting another
- **WHEN** the operator stops a single active profile
- **THEN** only that profile transitions to `disconnected` and unrelated profiles continue running

### Requirement: Port conflicts fail only the affected profile
The system SHALL detect local listener conflicts during startup and SHALL report the conflict on the affected profile without interrupting unrelated running tunnels.

#### Scenario: Second profile reuses an occupied local port
- **WHEN** the operator starts a profile whose requested local port is already bound by another running tunnel or process
- **THEN** that profile reports a startup failure and no unrelated active profile is stopped

### Requirement: Retry, keepalive, and diagnostics remain profile-scoped
The system SHALL apply retry logic, keepalive detection, and bounded log buffering independently for each running profile.

#### Scenario: Retry logs stay attached to the correct profile
- **WHEN** one profile encounters network timeouts and enters reconnect logic
- **THEN** the retry events appear only in that profile's diagnostics and do not pollute another profile's log stream
