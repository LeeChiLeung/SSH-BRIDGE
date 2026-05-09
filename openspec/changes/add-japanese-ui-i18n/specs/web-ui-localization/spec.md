## ADDED Requirements

### Requirement: New WebUI SHALL provide three-language localized text dictionaries
The system SHALL provide a centralized localization layer for the `chi + templ` WebUI that includes Simplified Chinese, English, and Japanese translations for shared navigation text, list labels, detail labels, form labels, action buttons, status labels, and empty-state text.

#### Scenario: Render a page in Japanese
- **WHEN** the operator selects Japanese in the new WebUI
- **THEN** the rendered page uses the Japanese dictionary for the UI-controlled text on that page

### Requirement: New WebUI SHALL render localized text through shared translation keys
The system SHALL render new WebUI text through shared translation keys rather than scattering hardcoded human-readable strings throughout handlers and templates.

#### Scenario: Reuse one translation key across multiple views
- **WHEN** the same UI action label appears in the list view, detail panel, and form modal
- **THEN** those views resolve the label through the same localization key instead of duplicating hardcoded text

### Requirement: New WebUI SHALL preserve a selected locale across page interactions
The system SHALL preserve the currently selected locale across page refreshes, detail navigation, form submissions, and other server-driven UI interactions within the new WebUI.

#### Scenario: Keep locale after form submission
- **WHEN** the operator changes the locale and then submits a form in the new WebUI
- **THEN** the response page remains in the selected locale instead of reverting to the default language
