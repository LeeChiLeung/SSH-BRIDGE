## ADDED Requirements

### Requirement: User can export tunnel profiles as JSON
The system SHALL allow the user to export all saved tunnel profiles from the new Web control console as a single JSON file.

#### Scenario: Export all profiles
- **WHEN** the user clicks the export button in the new control console
- **THEN** the system downloads a JSON file containing all saved tunnel profiles

#### Scenario: Export file includes schema metadata
- **WHEN** the system generates an exported profile file
- **THEN** the JSON file MUST include a schema version field and a profiles array

### Requirement: User can import tunnel profiles from JSON
The system SHALL allow the user to upload a JSON file from the new Web control console and validate it before any profile is written.

#### Scenario: Import file is parsed before write
- **WHEN** the user selects a JSON file for import
- **THEN** the system MUST parse and validate the file before saving any imported profile

#### Scenario: Invalid import file is rejected
- **WHEN** the uploaded file is not valid JSON or does not match the expected profile import schema
- **THEN** the system MUST reject the import and present an error instead of writing partial data

### Requirement: Import detects conflicts by tunnel name
The system SHALL detect profile conflicts by tunnel name using the profile title field from the import file and the current local profiles.

#### Scenario: Matching title is reported as a conflict
- **WHEN** an imported profile title matches the title of an existing local profile
- **THEN** the system MUST classify that imported profile as a conflict

#### Scenario: Missing title is rejected
- **WHEN** an imported profile has an empty title
- **THEN** the system MUST reject that imported profile as invalid

### Requirement: User can choose how conflicts are handled
The system SHALL require the user to choose a conflict policy before importing profiles whose names conflict with existing local profiles.

#### Scenario: User chooses skip for conflicts
- **WHEN** the import pre-check finds conflicts and the user chooses skip
- **THEN** the system MUST leave existing conflicting local profiles unchanged and continue importing only non-conflicting profiles

#### Scenario: User chooses merge for conflicts
- **WHEN** the import pre-check finds conflicts and the user chooses merge
- **THEN** the system MUST update the existing conflicting local profile using the imported profile fields while preserving the existing local profile identifier

### Requirement: Import result is summarized to the user
The system SHALL show the result of an import attempt, including how many profiles were added, skipped, merged, or rejected.

#### Scenario: Import summary after completion
- **WHEN** the system completes an import operation
- **THEN** the system MUST present counts for added, skipped, merged, and invalid profiles
