## ADDED Requirements

### Requirement: Local port forwarding through an SSH jump host
The tool SHALL open a local TCP listener on either a user-specified port or an automatically selected available port and SHALL forward traffic from each accepted local connection to a configured remote TCP endpoint through an SSH jump host.

#### Scenario: Forward traffic on a configured local port
- **WHEN** the user starts the tool with a valid SSH jump host, remote target, and explicit local port
- **THEN** the tool binds that local port and forwards TCP traffic bidirectionally between local clients and the remote target over SSH

#### Scenario: Forward traffic on an automatically selected port
- **WHEN** the user starts the tool without specifying a fixed local port or explicitly requests an ephemeral port
- **THEN** the tool binds an available local port, keeps that listener active for the lifetime of the process, and uses that bound port for forwarding

### Requirement: Resilient retry behavior for retryable SSH failures
The tool SHALL classify retryable SSH connection failures separately from fatal authentication failures and SHALL reconnect after retryable failures using exponential backoff starting at 1 second and capped at 60 seconds.

#### Scenario: Retry after a network timeout
- **WHEN** the SSH client fails to connect or loses an established session due to a network timeout, EOF, or connection reset
- **THEN** the tool retries connection establishment after a backoff delay of 1 second, doubling on consecutive failures up to a maximum of 60 seconds

#### Scenario: Reset backoff after recovery
- **WHEN** the tool successfully establishes a healthy SSH session after one or more retry attempts
- **THEN** the next retryable failure restarts the backoff sequence from 1 second

### Requirement: Fatal handling for authentication failures
The tool SHALL treat authentication failure as fatal and MUST exit without entering the reconnect loop.

#### Scenario: Stop on invalid credentials
- **WHEN** the SSH server rejects the supplied credentials during authentication
- **THEN** the tool reports an authentication failure and terminates instead of retrying

### Requirement: Active keep-alive detection of dead SSH sessions
The tool SHALL send active SSH keep-alive requests to the server and SHALL detect a dead or half-open SSH session within 15 seconds so that recovery logic can begin promptly.

#### Scenario: Detect a dead session
- **WHEN** the SSH connection becomes half-open or stops responding to keep-alive requests
- **THEN** the tool marks the session unhealthy within 15 seconds, closes the dead connection, and transitions into retry behavior

#### Scenario: Preserve a healthy session
- **WHEN** the SSH server responds successfully to periodic keep-alive requests
- **THEN** the tool keeps the session marked healthy and continues serving forwarded connections

### Requirement: Shell environment injection output
The tool SHALL support a CLI mode that prints a shell export command containing the effective local port to stdout for use with `eval $(...)`.

#### Scenario: Export the selected local port
- **WHEN** the user starts the tool with the environment export flag enabled
- **THEN** the tool prints `export TUNNEL_LOCAL_PORT=<bound-port>` to stdout after the listener is successfully bound

### Requirement: Graceful daemon-like shutdown
The tool SHALL continue running as a long-lived process until interrupted and SHALL handle `SIGINT` and `SIGTERM` by closing listeners, active tunnels, and background workers cleanly.

#### Scenario: Graceful shutdown on termination signal
- **WHEN** the process receives `SIGINT` or `SIGTERM`
- **THEN** the tool stops accepting new local connections, closes active forwarding connections, shuts down the SSH client, and exits without leaving the listener bound

### Requirement: Concurrent bidirectional stream forwarding
The tool SHALL use goroutines and `io.Copy` to forward bytes concurrently in both directions for each tunnelled TCP connection.

#### Scenario: Forward request and response streams simultaneously
- **WHEN** a local client and the remote target both send data on an active forwarded connection
- **THEN** the tool relays both streams concurrently without blocking one direction on activity in the other direction
