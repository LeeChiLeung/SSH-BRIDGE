## ADDED Requirements

### Requirement: The application can run a system tray entry on supported desktop platforms
The system SHALL provide a tray icon mode that can run on Windows, KDE-compatible Linux desktop environments, and macOS.

#### Scenario: Start tray mode on a supported platform
- **WHEN** the operator launches the application in tray mode on Windows, KDE/Linux, or macOS
- **THEN** the application creates a visible tray entry and remains running in the background

### Requirement: The tray menu exposes a console launcher and exit action
The system SHALL provide a tray menu with exactly two primary actions: opening the local console and exiting the application.

#### Scenario: Open the tray menu
- **WHEN** the operator opens the tray menu
- **THEN** the menu offers `打开控制台` and `退出`

#### Scenario: Exit from the tray menu
- **WHEN** the operator selects `退出`
- **THEN** the application shuts down cleanly and releases its tray and Web runtime resources

### Requirement: The tray entry does not introduce extra visual state in the first version
The system SHALL not expose tray notification badges, color changes, or other live status indicators in the first version.

#### Scenario: Review tray behavior
- **WHEN** the application is running in tray mode
- **THEN** the tray behaves as a static launcher and exit entry without status coloring or notification popups
