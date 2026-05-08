## Context

The project already has a working Go-based SSH tunneling runtime with reconnect logic, keepalive detection, and CLI startup. The next step is to make that runtime operable through a local browser console without replacing the current Go implementation or adding a SPA framework.

The requested UX is inspired by Stable Diffusion-style local control panels, but the provided sketch is more specific: a single-page list dashboard with a title header, a top-right `新增` action, and repeated tunnel cards that split information, state, and actions into separate visual blocks. The user explicitly does not want Vue, React, or similar frontend stacks, so the design must stay simple and server-driven.

## Goals / Non-Goals

**Goals:**
- Expose the tunnel runtime through a local HTTP server with control endpoints and status views.
- Provide a single-page local console using Go templates and HTMX for partial updates.
- Make the home page a card-based profile list rather than a single large configuration form.
- Allow operators to edit tunnel parameters, save named profiles, duplicate profiles, start and stop tunnels, and inspect runtime state and logs.
- Surface health information for both the local listener and the SSH control plane.
- Surface last enable time as part of the list-card summary data.
- Preserve the existing Go tunnel engine as the single source of tunnel behavior rather than duplicating logic for the Web layer.

**Non-Goals:**
- Introducing Vue, React, Svelte, or another SPA framework.
- Building a remote multi-user admin interface exposed to untrusted networks.
- Supporting multiple concurrently running tunnel instances in the first version.
- Adding user accounts, RBAC, or browser-based secrets vaulting.
- Embedding full verbose logs directly inside every list card.

## Decisions

Refactor the tunnel runtime into a reusable controller owned by the Go process.
Rationale: The Web UI needs start, stop, health, log, and state access without shelling out to the CLI. Moving the existing tunnel lifecycle into a shared controller lets both CLI and Web entrypoints invoke the same runtime and prevents drift in reconnect or shutdown behavior.
Alternatives considered:
- Launching the existing CLI as a subprocess from the Web server was rejected because it complicates process supervision, state reporting, and log streaming.

Serve the console from an embedded Go HTTP server with server-rendered templates plus HTMX.
Rationale: HTMX and templates provide a responsive local dashboard with low complexity, no build-heavy frontend stack, and easy integration with Go-owned state. This fits the user's explicit constraint against Vue/React while still allowing incremental updates for status cards, logs, and profile lists.
Alternatives considered:
- Plain full-page form posts were rejected because they make live state and log refresh clumsy.
- A React or Vue SPA was rejected per user requirement and would be disproportionate for a local admin console.

Use REST-style local endpoints for control and data retrieval.
Rationale: Distinct endpoints for `start`, `stop`, `status`, `health`, `logs`, `profiles`, and `duplicate` keep the server-side behavior explicit and testable. HTMX can consume these endpoints directly for partial page replacement, and future CLI or scripting integration can reuse them.

Model the WebUI home page as a card list with a separate detail region.
Rationale: The sketch shows each saved tunnel as a horizontal card with three logical zones: information, runtime state, and actions. This layout is better for scanning multiple tunnels than a single shared form. Logs and health details should live in a selected-card detail area or expandable panel, which preserves list readability while still exposing diagnostics.
Alternatives considered:
- A form-first dashboard was rejected because it does not match the sketch and scales poorly when many profiles exist.
- Putting recent logs directly inside each card was rejected because it makes the list visually noisy and reduces scan speed.

Persist saved tunnel profiles on local disk in a simple structured format.
Rationale: Operators need named reusable configurations, but the data model is modest. A local JSON file or equivalent simple serialization is enough for the first version and avoids introducing a database. Sensitive fields should avoid storing passphrases by default; key paths and non-secret tunnel parameters are the primary saved data. Each profile should also track metadata needed by the list view, including display title and last enable time.
Alternatives considered:
- SQLite was rejected for the first version because it adds operational surface without clear benefit.

Maintain an in-memory ring buffer for recent runtime logs.
Rationale: The console needs to show recent reconnect attempts, keepalive failures, authentication failures, and network timeouts without scraping stdout. A bounded in-memory log buffer avoids unbounded growth and provides immediate UI access.
Alternatives considered:
- Reading the process log file was rejected because it couples the UI to deployment-specific logging.

Bind the Web UI to loopback by default.
Rationale: The console controls network tunnels and displays key paths and connectivity details, so it should default to a local-only surface such as `127.0.0.1:<port>`.

## Risks / Trade-offs

- [Refactoring the CLI tunnel code into a shared controller may destabilize the working runtime] → Extract lifecycle boundaries carefully and keep the reconnect logic behaviorally identical to the current implementation.
- [Saved profiles can accidentally capture sensitive material] → Store key paths and connection metadata, but avoid persisting passphrases or raw private key contents.
- [HTMX polling for status and logs can become noisy] → Use small partial endpoints with bounded refresh intervals and bounded log buffers.
- [A card list can become cramped if too much detail is forced into the row] → Keep list cards focused on summary fields and move logs and deeper diagnostics into a selected-card detail region.
- [A single-process server and tunnel runtime means one crash affects both UI and tunnel] → Keep the control API thin and isolate runtime state transitions behind a controller abstraction.
- [Local-only binding may surprise users expecting LAN access] → Document the default and make listen address configurable if needed later.

## Migration Plan

Implementation will extend the current Go binary rather than replacing it. The rollout path is:
1. Extract the existing tunnel engine into a controller that supports programmatic lifecycle operations.
2. Add the HTTP server, card-list HTML templates, HTMX endpoints, and local profile persistence.
3. Keep CLI behavior working so existing terminal-driven usage remains available during the transition.

Rollback is straightforward: disable or remove the Web server entrypoint and continue using the CLI-only flow.

## Open Questions

- Should the detail region be an inline expandable panel under each card or a fixed lower workspace bound to the current selection?
- Should log updates be short-interval polling via HTMX, or should the first version expose server-sent events for smoother live updates?
