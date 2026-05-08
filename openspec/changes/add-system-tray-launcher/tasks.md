## 1. Add cross-platform tray mode

- [x] 1.1 Add a repository-supported tray mode entrypoint that can run on Windows, KDE/Linux, and macOS.
- [x] 1.2 Add tray menu actions for `打开控制台` and `退出`.
- [x] 1.3 Reuse existing shutdown behavior so exiting from the tray stops the local Web runtime cleanly.

## 2. Connect tray mode to the local Web console

- [x] 2.1 Start the local Web console on `127.0.0.1:8848` when running in tray mode.
- [x] 2.2 Implement default-browser launching for `http://127.0.0.1:8848` from the tray action.
- [x] 2.3 Keep tray mode resilient when browser launching fails, without adding notifications or tray state coloring.

## 3. Validate supported desktop behavior

- [x] 3.1 Verify that the tray menu shape matches the required actions on supported platforms.
- [x] 3.2 Verify that tray-triggered console launch targets `http://127.0.0.1:8848`.
- [x] 3.3 Verify that exiting from the tray stops the process cleanly while non-tray runtime paths remain intact.
