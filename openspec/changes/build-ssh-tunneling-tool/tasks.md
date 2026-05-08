## 1. CLI and process scaffolding

- [x] 1.1 Initialize the Go module entrypoint in `main.go` and add `golang.org/x/crypto/ssh` as a dependency.
- [x] 1.2 Define CLI flags for SSH jump host connection details, remote target address, local port selection, and `--env` output mode.
- [x] 1.3 Bind the local TCP listener at startup, support ephemeral port selection, and print `export TUNNEL_LOCAL_PORT=<bound-port>` only after a successful bind when `--env` is enabled.

## 2. SSH supervisor and retry lifecycle

- [x] 2.1 Implement the SSH connection state machine in `main.go` with clear comments for `connecting`, `connected`, retry, and shutdown transitions.
- [x] 2.2 Implement SSH dial logic that classifies authentication failures as fatal and network or timeout failures as retryable.
- [x] 2.3 Add capped exponential backoff starting at 1 second and doubling to a maximum of 60 seconds, with reset after a successful reconnect.
- [x] 2.4 Add active SSH keep-alive requests and mark the session unhealthy within 15 seconds when keep-alive probes fail.

## 3. Traffic forwarding and shutdown behavior

- [x] 3.1 Accept local TCP connections and open a corresponding remote TCP connection through the active SSH client for each accepted socket.
- [x] 3.2 Use goroutines and `io.Copy` to relay data bidirectionally and close both sides cleanly when either stream ends or errors.
- [x] 3.3 Handle temporary SSH unavailability by failing local connection attempts quickly instead of hanging accepted sockets.
- [x] 3.4 Trap `SIGINT` and `SIGTERM`, stop accepting new connections, close active tunnels and SSH state, and wait for worker shutdown before exiting.

## 4. Validation and operator ergonomics

- [ ] 4.1 Verify the tool against a reachable SSH server and remote TCP target for fixed-port forwarding, ephemeral-port forwarding, and `--env` shell injection behavior.
- [ ] 4.2 Verify retry and keep-alive behavior by simulating network interruption and confirming reconnect timing and dead-session detection.
- [ ] 4.3 Verify fatal exit on authentication failure and clean shutdown on termination signals.
