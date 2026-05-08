## ADDED Requirements

### Requirement: Tray actions open the local Web console in the default browser
The system SHALL open `http://127.0.0.1:8848` in the system default browser when the operator uses the tray console-launch action.

#### Scenario: Launch the console from the tray
- **WHEN** the operator clicks the tray action `打开控制台`
- **THEN** the application opens `http://127.0.0.1:8848` in the default browser

### Requirement: Tray mode runs the local Web console alongside the tray
The system SHALL ensure the local Web console is started when the application runs in tray mode so the launched browser target is available locally.

#### Scenario: Start tray mode from a packaged desktop app
- **WHEN** the operator launches the desktop-oriented application entrypoint in tray mode
- **THEN** the local Web console is started and listens on `127.0.0.1:8848` before or as the tray becomes usable

### Requirement: Browser launch failures remain recoverable
The system SHALL keep the tray process running even if a browser-open attempt fails, so the operator can retry or exit cleanly.

#### Scenario: Default browser open fails
- **WHEN** the operator triggers `打开控制台` but the browser open action fails
- **THEN** the tray process remains active and the operator can continue using tray actions including `退出`
