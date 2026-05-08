## 1. Add runtime traffic accounting

- [x] 1.1 Extend profile snapshots with traffic fields for current upload/download speed, cumulative upload/download bytes, and active connection count.
- [x] 1.2 Add concurrency-safe per-profile counters and a periodic sampler that derives current BPS values while a tunnel runtime is active.
- [x] 1.3 Count uploaded and downloaded bytes inside the forwarding path and keep active connection counts accurate across connect, retry, and stop transitions.

## 2. Expose metrics through the Web console

- [x] 2.1 Update profile card rendering to show text-based traffic metrics for each profile snapshot.
- [x] 2.2 Update detail and diagnostics rendering to show the selected profile's current speeds, cumulative totals, and active connection count.
- [x] 2.3 Include traffic fields in the Web status API payloads that return profile snapshots.

## 3. Validate profile-scoped metrics behavior

- [x] 3.1 Verify that traffic through one active profile does not affect the counters shown for another active profile.
- [x] 3.2 Verify that upload/download totals increase when data is forwarded and that current speed returns to zero when traffic becomes idle or the profile stops.
- [x] 3.3 Verify that active connection counts rise while forwarded connections are open and return to zero after disconnect or stop.
