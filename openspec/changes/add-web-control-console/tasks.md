## 1. Refactor tunnel control into a shared runtime

- [x] 1.1 Extract the current tunnel lifecycle logic from the CLI-only flow into a reusable controller with start, stop, status, and health accessors.
- [x] 1.2 Add a bounded in-memory log buffer so reconnect, keepalive, authentication, and network events can be read by the Web layer.
- [x] 1.3 Preserve the existing CLI startup path by making it use the shared controller instead of a separate tunnel code path.

## 2. Add the local HTTP control surface

- [x] 2.1 Add a local-only Go HTTP server that serves the Web console and exposes endpoints for start, stop, status, health, logs, and profile management.
- [x] 2.2 Define request and response shapes that cover SSH address, SSH port, SSH user, local bind address, local port, remote target, and private key path.
- [x] 2.3 Ensure the start and stop endpoints drive the shared tunnel controller and return updated runtime state for UI refreshes.
- [x] 2.4 Add profile duplication support in the local control API.

## 3. Build the server-rendered Web console

- [x] 3.1 Add Go HTML templates for a single-page local dashboard with a title header, a top-level `新增` action, and a card-based saved profile list.
- [x] 3.2 Render each tunnel profile card with separate information, runtime state, and action regions, including SSH address, local proxy port, remote port, and last enable time.
- [x] 3.3 Add HTMX-driven partial updates for runtime state, card health indicators, and recent detail-view log entries without introducing Vue, React, or another SPA framework.
- [x] 3.4 Add UI behavior for creating, editing, and duplicating profiles, plus showing active endpoint details clearly.
- [x] 3.5 Add a detail region or expandable panel for logs and deeper health diagnostics so list cards remain concise.

## 4. Persist profiles and expose health data

- [x] 4.1 Add local persistence for named tunnel profiles using a simple on-disk format suitable for single-user local use.
- [x] 4.2 Avoid persisting passphrases or raw private key contents while still saving reusable non-secret tunnel settings and key paths.
- [x] 4.3 Persist and update profile metadata needed by the list view, including display title and last enable time.
- [x] 4.4 Implement health reporting that distinguishes local listener readiness from SSH control-plane connectivity.

## 5. Validate the WebUI workflow

- [x] 5.1 Verify that profiles appear as cards on the home page and can be created from the `新增` action.
- [x] 5.2 Verify that a tunnel can be started, observed, stopped, edited, and duplicated entirely from the browser console.
- [x] 5.3 Verify that runtime state transitions show `disconnected`, `connecting`, `connected`, and `retrying` correctly in the card UI.
- [x] 5.4 Verify that saved profiles persist across restarts, that last enable time updates correctly, and that the detail log and health panels reflect real tunnel events.
