## Context

This change introduces a standalone Go CLI that establishes a local TCP listener and forwards each accepted connection to a remote target through an SSH jump server. The tool is intended for production-like operational use, so the design must prioritize connection resilience, predictable shutdown, and clear operator behavior over minimal implementation size.

The implementation is constrained to Go 1.20+, `golang.org/x/crypto/ssh`, and a single `main.go` containing the core logic. The system must distinguish fatal authentication errors from retryable transport failures, support both fixed and ephemeral local ports, and expose the selected port through shell-friendly environment export output.

## Goals / Non-Goals

**Goals:**
- Provide local TCP port forwarding through a single SSH jump host to an internal TCP endpoint.
- Maintain service continuity by reconnecting with exponential backoff after retryable SSH failures.
- Detect half-open or dead SSH sessions within 15 seconds using active keep-alive probes.
- Allow shell workflows to discover the effective bound port through `--env` output.
- Shut down cleanly on `SIGINT` and `SIGTERM`, closing listeners, active sessions, and worker goroutines.
- Keep the implementation understandable in one `main.go`, with comments around the connection state machine.

**Non-Goals:**
- Supporting remote port forwarding, dynamic SOCKS proxying, or multi-hop SSH chains.
- Managing credentials beyond consuming them from CLI arguments, environment variables, or key files.
- Providing a supervisor, init script, or OS-specific service integration.
- Multiplexing multiple distinct tunnel definitions in one process.

## Decisions

Use an explicit tunnel supervisor state machine in `main.go`.
Rationale: The reconnect, health-check, and shutdown behavior is the core complexity of this tool. A small state machine with states such as `disconnected`, `connecting`, `connected`, and `stopping` makes retry decisions and fatal-exit conditions auditable. This is preferable to scattering reconnect logic across accept loops and per-connection handlers.
Alternatives considered:
- A purely linear loop with ad hoc booleans was rejected because it obscures why the process exits versus retries.
- A larger multi-package architecture was rejected because the deliverable explicitly favors a single-file implementation.

Keep the local listener stable across SSH reconnects.
Rationale: Binding the local port once avoids breaking callers that have already learned the port, especially when `--env` is used with an ephemeral port. The listener can continue accepting local clients while the current SSH session status determines whether each connection is forwarded immediately or rejected quickly.
Alternatives considered:
- Rebinding on every reconnect was rejected because it can change the port and break automation.

Use one SSH client connection as the control plane and open a fresh remote TCP connection per accepted local connection.
Rationale: This matches standard local port forwarding semantics. Each accepted local socket maps to a `client.Dial("tcp", remoteAddr)` call over the active SSH client, then two `io.Copy` goroutines provide bidirectional streaming until either side closes.
Alternatives considered:
- Reusing a single remote channel for multiple local clients was rejected because it does not match TCP forwarding behavior and complicates isolation.

Implement keep-alive with `client.Conn.SendRequest("keepalive@openssh.com", true, nil)` on a periodic ticker.
Rationale: Active requests surface dead sessions that would otherwise remain half-open at the TCP layer. A 5-second interval with failure threshold or deadline bounded to 15 seconds satisfies the requirement to detect dead connections quickly without waiting for application traffic.
Alternatives considered:
- Passive TCP keepalive alone was rejected because it is slower and less predictable across environments.

Classify errors before retrying.
Rationale: Authentication failures are operator-actionable misconfiguration and must terminate immediately. Network timeouts, EOFs, connection resets, and keep-alive failures are transient and should transition back to exponential backoff retry.
Alternatives considered:
- Treating all dial failures as retryable was rejected because invalid credentials would otherwise spin forever.

Use capped exponential backoff with reset on successful connection establishment.
Rationale: Backoff values of 1s, 2s, 4s and so on capped at 60s provide quick recovery for short interruptions without creating hot retry loops during outages. Resetting the backoff after a healthy connection ensures future failures start with a short retry interval.

Emit `export TUNNEL_LOCAL_PORT=<port>` to stdout when `--env` is set.
Rationale: This supports `eval $(tool --env ...)` in POSIX shells. The command output should remain machine-friendly and avoid extra prose in this mode.
Alternatives considered:
- Writing dotenv files was rejected because the requirement is shell-context injection, not file generation.

## Risks / Trade-offs

- [Accept loop receives local connections while SSH is temporarily down] → Reject the connection quickly with a logged error instead of blocking until reconnect, so callers fail fast and can retry.
- [Single-file implementation can become hard to maintain] → Use small structs and functions inside `main.go`, plus targeted comments around the supervisor and lifecycle transitions.
- [Keep-alive frequency increases SSH control traffic] → Use lightweight global requests at a modest interval; the reliability gain outweighs the negligible traffic cost.
- [Ephemeral port output can race with process startup if printed too early] → Bind the listener before printing `--env` output so the reported port is authoritative.
- [Graceful shutdown can leak goroutines if copies block forever] → Close the listener, SSH client, and active local/remote connections during shutdown so `io.Copy` returns.

## Migration Plan

No data migration is required. Implementation rollout consists of adding `main.go`, wiring module dependencies, and validating the binary against a reachable SSH jump host and a known internal TCP service.

Rollback is straightforward: remove the new binary or revert the commit introducing the tunnel tool if the behavior does not meet operational requirements.

## Open Questions

- Should CLI parameters use only flags, or should selected credentials also be readable from environment variables for non-interactive automation?
- Should the process log in plain text only, or is a structured log format desired for production deployments?
