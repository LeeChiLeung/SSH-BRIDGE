## Why

The project currently allows saving many tunnel profiles, but only one tunnel runtime can be active at a time because the controller owns a single running instance. This blocks the tool from serving common operational use cases where multiple SSH forwards need to stay up simultaneously and makes the card-based Web console misleading when several profiles appear runnable but only one can actually be active.

## What Changes

- Upgrade the tunnel runtime from a single active instance model to a multi-instance tunnel manager.
- Allow multiple saved tunnel profiles to run at the same time, with each profile maintaining its own listener, SSH connection lifecycle, health state, and logs.
- Change WebUI status rendering from one global active runtime to per-profile runtime state.
- Allow start and stop actions to affect only the targeted profile rather than replacing the currently running tunnel.
- Preserve profile-level diagnostics, endpoint display, enable time updates, and retry behavior independently for each running tunnel.
- Keep the existing CLI tunnel behavior working while introducing a manager abstraction that can host one or many tunnel runs.

## Capabilities

### New Capabilities
- `multi-instance-tunnel-manager`: Ability to run and monitor multiple independent SSH tunnel instances concurrently within the same application.
- `multi-instance-web-console`: Per-profile Web console state and actions for concurrently running tunnels.

### Modified Capabilities
- None.

## Impact

- Refactors the current single `TunnelController` runtime model into a manager of per-profile tunnel runs.
- Changes the WebUI data model, rendering logic, and APIs from global runtime state to profile-scoped runtime state.
- Requires careful port-conflict handling and per-instance lifecycle isolation.
- Preserves the current SSH forwarding engine, retry logic, keepalive behavior, and logging model, but applies them per instance.
