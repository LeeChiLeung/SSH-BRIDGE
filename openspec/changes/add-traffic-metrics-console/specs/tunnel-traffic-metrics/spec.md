## ADDED Requirements

### Requirement: Tunnel runtime tracks traffic totals per profile
The system SHALL track cumulative uploaded bytes, cumulative downloaded bytes, and active forwarded connection count independently for each running tunnel profile.

#### Scenario: Two active profiles carry different traffic volumes
- **WHEN** two tunnel profiles are running concurrently and traffic flows through both
- **THEN** each profile reports its own upload total, download total, and active connection count without mixing values from the other profile

#### Scenario: Active connections change during forwarding
- **WHEN** a local client connection is accepted and later closed for a running tunnel profile
- **THEN** that profile's active connection count increases while the connection is live and decreases again after the forwarding pair is torn down

### Requirement: Tunnel runtime reports current transfer speed per profile
The system SHALL calculate current upload and download speed for each running profile using periodic sampling of forwarded byte counters.

#### Scenario: Traffic is flowing through a connected tunnel
- **WHEN** bytes are being forwarded over a running profile during a sampling interval
- **THEN** the profile reports non-zero current upload and/or download speed for the corresponding transfer direction

#### Scenario: Traffic becomes idle
- **WHEN** a running profile remains connected but no bytes are forwarded during a sampling interval
- **THEN** the reported current upload and download speed fall back to zero for that interval

### Requirement: Traffic metrics reset live activity correctly on shutdown or failure
The system SHALL stop reporting non-zero live speed and active connection counts for a profile once that runtime is no longer forwarding traffic.

#### Scenario: A profile is stopped by the operator
- **WHEN** the operator stops an active profile
- **THEN** the profile reports zero active connections and zero current upload/download speed after shutdown completes

#### Scenario: A profile enters retrying after connection loss
- **WHEN** the SSH session drops and the profile transitions into retry logic
- **THEN** the profile reports zero active connections and zero current upload/download speed until forwarding resumes again
