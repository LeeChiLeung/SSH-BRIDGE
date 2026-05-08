## Why

The SSH tunneling tool now has working CLI-based forwarding, but operating it through terminal flags alone is awkward for repeated daily use, status inspection, and managing multiple reusable tunnel definitions. A local Web control console is needed to provide a list-oriented operational dashboard where saved tunnels can be scanned, started, stopped, copied, and inspected quickly without introducing a heavyweight frontend stack.

## What Changes

- Add a local Web control console whose home page is a card-based list of saved tunnel profiles.
- Add a top-level page layout with a title area, a `新增` action, and a main list workspace for tunnel cards.
- Add card-based tunnel summaries that show SSH address, local proxy port, remote port, and last enable time.
- Add per-card tunnel lifecycle controls for start and stop without requiring manual CLI invocation.
- Add per-card actions for edit and duplicate so operators can adjust or clone existing profiles quickly.
- Add real-time tunnel state display on each card covering `disconnected`, `connecting`, `connected`, and `retrying`.
- Add a log and health detail area associated with a selected or expanded tunnel rather than forcing full logs into every list card.
- Add current port and endpoint display so operators can immediately see the active local entrypoint for each running tunnel.
- Add health indicators for local listener readiness and SSH control-plane connectivity.
- Add persisted tunnel profiles so frequently used configurations can be saved and reused.
- Keep the implementation server-rendered with Go templates and HTMX rather than introducing Vue, React, or other SPA frameworks.

## Capabilities

### New Capabilities
- `web-control-console`: Local browser-based control plane for configuring, running, monitoring, and saving SSH tunnel profiles.

### Modified Capabilities
- None.

## Impact

- Extends the Go application with an embedded HTTP server, REST-style control endpoints, template rendering, and HTMX-driven updates for list cards and detail panels.
- Requires refactoring the current tunnel runtime so it can be controlled programmatically by both the CLI and Web layers.
- Introduces local persistence for saved tunnel profiles and active runtime status.
- Adds a new operator-facing list dashboard while preserving the existing Go-based tunnel engine.
