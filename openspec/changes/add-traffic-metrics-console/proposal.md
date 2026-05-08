## Why

Operators need visibility into whether a tunnel is merely connected or actually carrying useful traffic. The current console shows state and logs, but it cannot answer practical questions such as current upload/download speed, cumulative transfer volume, or how many client connections are active on each tunnel.

## What Changes

- Add per-profile traffic accounting for upload bytes, download bytes, and active forwarded connection count.
- Add rolling transfer-rate calculation so each running tunnel reports current upload and download speed.
- Surface traffic metrics through the runtime snapshot and HTTP APIs alongside existing state, health, and logs.
- Extend the Web console card and detail views to display the metrics as text, without adding charts or graphical widgets in the first version.
- Keep metrics isolated per running profile so concurrent tunnels do not mix traffic counters.

## Capabilities

### New Capabilities
- `tunnel-traffic-metrics`: Per-profile traffic accounting, live throughput, and active connection tracking for SSH tunnel runtimes
- `traffic-metrics-web-console`: Text-based Web console presentation of transfer metrics for each tunnel profile

### Modified Capabilities
- None.

## Impact

- Affects the tunnel runtime in `controller.go`, especially connection proxying and profile snapshots.
- Affects Web rendering and API payloads in `web.go`.
- Requires additional concurrency-safe counters and periodic sampling logic, but no external dependencies.
