## Context

The current implementation has evolved from a CLI-only SSH tunnel into a local Web-managed console, but the runtime model is still single-instance. A single `TunnelController` owns exactly one `tunnelRun`, one state value, one health snapshot, and one log buffer, which means starting one profile displaces any previously running profile.

This creates a mismatch with the product direction: the application already stores multiple tunnel profiles and renders them as separate cards, but only one card can actually represent a live runtime. The multi-instance change must preserve the existing SSH forwarding behavior while making runtime state, logs, and health independent per profile.

## Goals / Non-Goals

**Goals:**
- Allow multiple tunnel profiles to run concurrently in the same application process.
- Preserve independent listener binding, SSH retry state, keepalive monitoring, endpoint reporting, and log buffers per running profile.
- Update the Web console so each card controls and displays only its own runtime.
- Preserve the existing CLI behavior, while implementing it on top of the same multi-instance manager abstraction.
- Handle profile-specific failures without disrupting unrelated running tunnels.

**Non-Goals:**
- Coordinating tunnels across multiple machines or processes.
- Sharing one SSH client across different profiles.
- Solving conflicting local port assignments by auto-rewriting ports in the first version.
- Adding distributed state, clustering, or remote orchestration.

## Decisions

Replace the single `TunnelController` runtime slot with a manager keyed by profile ID.
Rationale: The fundamental limitation today is structural. A multi-instance feature requires a stable lookup of runtime state by profile ID, so the application needs `map[profileID]*tunnelRun` plus per-profile snapshots instead of one global `run`.
Alternatives considered:
- Spawning one controller process per profile from the Web server was rejected because it complicates lifecycle management and makes per-card state aggregation harder.

Keep one isolated SSH session lifecycle per running profile.
Rationale: Each tunnel profile can have a different jump host, local port, or remote target. Isolating retries, keepalives, listeners, and log buffers per profile prevents one profile's network instability from corrupting the runtime state of another.
Alternatives considered:
- Sharing a single SSH client across profiles was rejected because profile settings and failure domains are independent.

Model WebUI state and APIs per profile rather than global.
Rationale: The current Web console exposes one global runtime snapshot. That must become profile-scoped so each card can render its own `disconnected`, `connecting`, `connected`, or `retrying` state, endpoint, health, and logs.
Alternatives considered:
- Keeping one global `/api/status` was rejected because it cannot represent concurrent active runs accurately.

Treat local port conflicts as profile-specific startup failures.
Rationale: Concurrent tunnels only work if each profile can bind its own local port. When a port is already in use, the failing profile should surface a start error without impacting already running tunnels.
Alternatives considered:
- Automatically assigning a random replacement port was rejected because it changes operator intent and breaks predictable local endpoints.

Retain per-profile bounded log buffers.
Rationale: Diagnostics need to remain attributable to the right card and detail panel. A shared log stream would become noisy and would not support clean per-profile diagnostics.

## Risks / Trade-offs

- [More concurrent goroutines and listeners increase runtime complexity] → Keep each `tunnelRun` self-contained and managed through a narrow manager API.
- [Web handlers can race with concurrent runtime updates] → Use manager-level locking and immutable snapshot copies for rendered views.
- [Port conflicts will become more visible once users run multiple profiles] → Surface explicit profile-scoped bind errors on the affected card.
- [CLI semantics may become ambiguous if multiple runs already exist] → Keep CLI direct-run mode scoped to a named runtime slot and document that it starts or replaces only its own slot.
- [Per-profile polling can multiply request volume] → Restrict refreshes to selected detail views and explicit card refreshes, not whole-page hot loops.

## Migration Plan

1. Refactor the runtime into a multi-instance manager while keeping current single-profile behavior working through the manager interface.
2. Move status, health, endpoint, and logs from global state to profile-scoped snapshots.
3. Update Web handlers and templates so cards query and mutate profile-scoped runtime state.
4. Validate concurrent runs using at least two profiles with distinct local ports.

Rollback is straightforward: revert to the single-instance controller if the manager abstraction introduces regressions.

## Open Questions

- Should the CLI support naming arbitrary runtime slots directly, or is profile ID enough for the first version?
- Should the home page sort active profiles before inactive ones once many tunnels can run simultaneously?
