## Context

The current multi-instance runtime already maintains per-profile snapshots for state, endpoint, health, errors, and logs. Traffic visibility is the next operational layer: the application needs to measure bytes flowing through the two `io.Copy` directions and expose those numbers on the same per-profile snapshot model without coupling different tunnel profiles together.

The existing forwarding model is connection-oriented and concurrent. Each accepted local connection creates a remote SSH-backed TCP connection and starts two goroutines that stream bytes in opposite directions. This makes traffic metrics straightforward to observe at the proxy boundary, but all counters must be safe under concurrent access and cheap enough to avoid degrading forwarding performance.

## Goals / Non-Goals

**Goals:**
- Track upload bytes, download bytes, and active forwarded connections per running profile.
- Calculate current upload and download speed at a fixed cadence suitable for human-readable Web display.
- Expose metrics through profile snapshots so cards, details, and APIs can render them consistently.
- Keep traffic counters isolated so one tunnel's load never contaminates another tunnel's metrics.

**Non-Goals:**
- Packet-level, SSH-encrypted wire-size accounting.
- Historical persistence of traffic metrics after a tunnel stops or the process exits.
- Graphs, sparklines, or other visualizations beyond text-based rendering.
- Per-remote-host or per-client breakdowns inside a tunnel profile.

## Decisions

Add a profile-scoped `TrafficStats` structure to tunnel snapshots.
Rationale: The Web console and APIs already consume snapshot state per profile, so traffic data should join that contract rather than living in a separate query path.
Alternatives considered:
- A separate metrics map outside the snapshot model was rejected because it would complicate Web rendering and increase consistency risks.

Count bytes inside the proxy loop by wrapping the `io.Copy` source side with a counting reader.
Rationale: The tunnel already moves traffic through `io.Copy`; counting at that boundary captures the payload the application is actually forwarding and keeps the implementation local to the runtime path.
Alternatives considered:
- Counting bytes on listeners or raw sockets was rejected because it would blur directionality and add more invasive plumbing.

Use atomic counters for cumulative bytes and active connections, plus a lightweight sampling loop for bytes-per-second.
Rationale: Connection proxy goroutines update counters frequently. Atomic increments avoid coarse locks on the hot path, while a per-profile sampler can compute rates from byte deltas at one-second intervals.
Alternatives considered:
- Locking every read/write update was rejected due to avoidable contention.
- Computing speed directly in each `Read` call was rejected because rate calculation belongs in a lower-frequency observer, not the hot path.

Reset live speed and active-connection indicators when a profile stops, but keep the last cumulative totals until that runtime snapshot is replaced.
Rationale: Operators usually want totals for the current run even after a disconnect, while stale non-zero BPS or connection counts would be misleading.
Alternatives considered:
- Clearing all counters immediately on stop was rejected because it removes useful diagnostics right after a failure.

Render metrics as text in existing card and detail regions.
Rationale: The request explicitly prefers a pragmatic first version without front-end framework expansion or charting work.
Alternatives considered:
- Separate dashboards or visual charts were rejected as unnecessary scope for the first increment.

## Risks / Trade-offs

- [Traffic counters add overhead on every proxied stream] → Keep hot-path accounting to atomic increments only and avoid allocation-heavy wrappers.
- [Instantaneous BPS values can appear noisy for bursty traffic] → Use a fixed sampling interval and present values as current speed rather than long-term averages.
- [Stopped tunnels may show cumulative bytes from the last run] → Document that totals are per-runtime snapshot and reset on the next start for that profile.
- [Web polling can show slightly stale rates] → Reuse the existing manual/optional refresh model; exact real-time precision is not required.

## Migration Plan

1. Extend runtime snapshots with traffic fields and wire counters into the proxy path.
2. Add per-profile sampler logic that updates current upload/download BPS while the profile is running.
3. Expose the fields through Web card/detail rendering and JSON APIs.
4. Verify the counters with local tests that generate traffic over a live tunnel.

Rollback is straightforward: remove the new snapshot fields and hot-path counting code without affecting the core forwarding behavior.

## Open Questions

- Should cumulative byte counters reset automatically when a stopped profile is started again, or should the application retain an all-time per-profile total in a future change?
