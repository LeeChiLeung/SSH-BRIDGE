## 1. Refactor the runtime into a multi-instance manager

- [x] 1.1 Replace the single active `TunnelController` run slot with a manager that tracks running tunnel instances by profile ID.
- [x] 1.2 Keep each running profile isolated with its own listener, SSH session lifecycle, cancelation context, and tracked connections.
- [x] 1.3 Preserve the existing SSH retry and keepalive behavior while applying it independently per running profile.

## 2. Isolate state, health, and diagnostics per profile

- [x] 2.1 Move runtime snapshots from one global state object to profile-scoped snapshots containing state, endpoint, health, last error, and recent logs.
- [x] 2.2 Keep bounded log buffers per running profile so diagnostics remain attributable to the correct tunnel card.
- [x] 2.3 Surface local port bind conflicts and startup failures only on the affected profile without interrupting unrelated active tunnels.

## 3. Update CLI and Web control flows for concurrent runs

- [x] 3.1 Update direct CLI tunnel startup to use the multi-instance manager without reintroducing a separate runtime path.
- [x] 3.2 Update Web start and stop handlers so card actions affect only the selected profile.
- [x] 3.3 Update card rendering to use per-profile runtime state instead of one global active snapshot.
- [x] 3.4 Update detail and diagnostics handlers so they render only the selected profile's endpoint, health, errors, and logs.

## 4. Validate concurrent multi-profile behavior

- [x] 4.1 Verify that two profiles with distinct local ports can be started and remain active at the same time.
- [x] 4.2 Verify that stopping one active profile does not stop the other profile.
- [x] 4.3 Verify that one profile can enter `retrying` while another remains `connected`, and that the WebUI shows those states on the correct cards.
- [x] 4.4 Verify that a local port conflict fails only the affected profile and leaves unrelated active profiles running.
