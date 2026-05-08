## Why

Teams often need temporary or long-lived access to private TCP services behind an SSH jump host, but ad hoc port forwarding breaks down under unstable network conditions and does not integrate cleanly with automation. A production-ready tunneling tool is needed now to provide resilient forwarding, shell-friendly environment injection, and predictable shutdown behavior in a single Go binary.

## What Changes

- Add a standalone Go CLI that creates local TCP listeners and forwards traffic to a remote internal target through an SSH jump server.
- Support binding either an explicit local port or an automatically selected ephemeral port.
- Add resilient SSH session management with exponential backoff reconnect behavior for retryable network failures.
- Add active SSH keep-alive probing to detect dead or half-open connections within 15 seconds and trigger recovery.
- Add an `--env` mode that prints shell export commands so callers can inject the selected local port into their environment with `eval $(...)`.
- Add graceful signal handling for `SIGINT` and `SIGTERM` so listeners, active connections, and background loops are closed cleanly.
- Enforce distinct handling for fatal authentication errors versus retryable network and timeout failures.

## Capabilities

### New Capabilities
- `ssh-tunnel-cli`: Production-grade SSH local port forwarding CLI with resilient connection management and environment export support.

### Modified Capabilities
- None.

## Impact

- Adds a new Go entrypoint in `main.go` with SSH dialing, listener lifecycle management, signal handling, and bidirectional stream forwarding.
- Introduces a runtime dependency on `golang.org/x/crypto/ssh`.
- Defines CLI behavior for port selection, target selection, environment export output, and process lifecycle.
- Establishes the baseline spec and implementation plan for a background-capable SSH tunneling binary.
