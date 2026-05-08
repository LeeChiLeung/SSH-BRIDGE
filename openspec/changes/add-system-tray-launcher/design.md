## Context

The application already has a local Web console and a release packaging flow. What it lacks is a desktop shell that keeps the service resident, exposes a simple tray entry point, and opens the browser for the operator. This is especially useful on Windows and macOS, where background apps are expected to provide a launcher surface, and on KDE where tray behavior is generally supported well.

The tray feature is a desktop integration concern, not a tunnel-runtime concern. It should wrap the existing Web console startup and browser launch behavior without changing forwarding logic, profile storage, or the console's HTTP API.

## Goals / Non-Goals

**Goals:**
- Provide a cross-platform tray icon entry for the application.
- Make the tray menu open the local Web console at `http://127.0.0.1:8848`.
- Provide a simple exit action from the tray.
- Support Windows, KDE-friendly Linux desktop environments, and macOS.
- Keep the feature minimal and predictable for the first release.

**Non-Goals:**
- Tray notification badges or live state colors.
- Multiple windows or embedded browser UI.
- Custom in-app browser rendering.
- Deep desktop-service integration such as auto-start registration or update notifications.

## Decisions

Use a single tray-driven desktop mode layered on top of the existing Web console.
Rationale: The product already uses a browser-based control console, so the tray only needs to provide a stable launcher and exit path.
Alternatives considered:
- Embedding the UI into a native window was rejected because it would duplicate the existing Web console UI and complicate maintenance.

Open the system default browser rather than shipping an embedded browser.
Rationale: The console already runs locally in a browser-friendly form, and relying on the user's default browser reduces platform-specific UI complexity.
Alternatives considered:
- Embedding WebView was rejected because it would introduce more platform-specific surface area and packaging complexity.

Keep tray behavior minimal and menu-driven.
Rationale: The first version should focus on reliable launch/exit behavior. Status indicators and rich menu state can be added later once the core desktop shell is stable.
Alternatives considered:
- Live health indicators and colored tray states were rejected for the first cut because they require additional runtime wiring and platform nuance.

Use a cross-platform tray library rather than separate native tray implementations.
Rationale: The same menu semantics are needed across Windows, KDE/Linux, and macOS, and a shared abstraction keeps the codebase smaller.
Alternatives considered:
- Hand-writing OS-specific tray code for each platform was rejected because it would fragment the implementation and increase maintenance cost.

## Risks / Trade-offs

- [Tray support differs across desktop environments] → Target KDE-friendly Linux environments first and keep the tray API simple enough to degrade gracefully where system tray support is limited.
- [macOS background app packaging has extra constraints] → Keep the tray feature aligned with the existing AppImage/release strategy while documenting that macOS packaging may need `.app` wrapping later.
- [Opening the browser may fail if no default browser is configured] → Surface a simple error path and keep the console URL visible for manual access.
- [Tray lifecycle bugs can prevent clean shutdown] → Treat exit as a first-class menu action and reuse the existing signal/shutdown logic for cleanup.

## Migration Plan

1. Introduce a desktop mode that starts the local Web console and tray entry together.
2. Wire the tray menu to the default browser open action and exit action.
3. Verify behavior on Windows, KDE/Linux, and macOS packaging targets.
4. Keep non-tray CLI mode unchanged for users who still prefer direct tunnel startup.

Rollback is straightforward: disable the tray entrypoint and continue using the Web console directly.

## Open Questions

- Should the tray mode become the default desktop launch path for all packaged releases, or only for AppImage and Windows/macOS bundles?
