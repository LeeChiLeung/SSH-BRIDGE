## ADDED Requirements

### Requirement: Web console SHALL expose a visible language switcher in the new UI
The system SHALL expose a visible language switcher in the new `chi + templ` control console so that the operator can switch between Simplified Chinese, English, and Japanese without editing URLs or restarting the application.

#### Scenario: Switch locale from the console header
- **WHEN** the operator clicks a language switch control in the new console
- **THEN** the application updates the current UI locale and reloads or refreshes the console in the selected language

### Requirement: Web console SHALL keep the selected locale during embedded detail workflows
The system SHALL keep the selected locale while the operator navigates between the list view, embedded detail panel, modal forms, and related redirects in the new console.

#### Scenario: Keep locale while opening and closing edit modal
- **WHEN** the operator selects a non-default language, opens an edit modal, and closes or submits it
- **THEN** the console returns to the same workflow state in the same selected language
