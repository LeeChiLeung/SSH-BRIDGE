## Why

The application has evolved into a local desktop-style service with a browser-based control console, but it still launches like a CLI-first tool. A system tray entry is needed now so users on KDE, Windows, and macOS can keep the service running in the background and reopen the local console without manually managing terminal windows or remembering URLs.

## What Changes

- Add a cross-platform system tray mode that can run alongside the local Web console.
- Make tray interaction open `http://127.0.0.1:8848` in the system default browser.
- Add a minimal tray menu with `打开控制台` and `退出`.
- Support the tray feature on Windows, KDE-compatible Linux desktop environments, and macOS.
- Keep the first version intentionally simple by not adding tray status badges, color changes, or notification flows.

## Capabilities

### New Capabilities
- `cross-platform-system-tray`: Cross-platform tray icon lifecycle and menu behavior for the desktop-hosted application
- `tray-web-console-launch`: Tray-triggered launch of the local Web console in the default browser

### Modified Capabilities
- None.

## Impact

- Adds a desktop integration layer and likely a cross-platform tray dependency.
- Affects application startup flow, local Web mode entrypoints, and packaged desktop behavior.
- Does not change tunnel forwarding logic or the Web console feature set itself.
