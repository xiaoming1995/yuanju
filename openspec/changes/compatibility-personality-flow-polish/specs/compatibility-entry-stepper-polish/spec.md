## ADDED Requirements

### Requirement: Compatibility entry shows compact consultation progress
The compatibility creation page SHALL present the flow as a compact consultation sequence without turning it into a blocking wizard.

#### Scenario: User opens compatibility entry
- **WHEN** a user opens the compatibility creation page
- **THEN** the page shows visible progress cues for selected question, selected relationship stage, and birth-profile completion
- **AND** all required fields remain editable on the same page

#### Scenario: User changes context selections
- **WHEN** the user changes the primary question or relationship stage
- **THEN** the progress cue and preview copy update to reflect the selected context

### Requirement: Birth profile forms remain reachable after the consultation cue
The compatibility creation page SHALL keep both birth profile forms easy to reach after the personality consultation controls.

#### Scenario: Desktop layout
- **WHEN** the entry page is viewed on desktop
- **THEN** question/stage controls appear before the birth-profile forms
- **AND** the two birth-profile forms remain visible or clearly reachable without a modal

#### Scenario: Mobile layout
- **WHEN** the entry page is viewed on a narrow viewport
- **THEN** progress cues, active profile tabs, active birth form, and primary action stack without overlapping the bottom navigation

### Requirement: Entry copy remains personality-first
The compatibility creation page SHALL frame the reading around personality fit and relationship interaction rather than raw scoring.

#### Scenario: User reviews the first screen
- **WHEN** the user reads the entry-page introduction and preview
- **THEN** the page emphasizes personality fit, fit points, conflict points, and interaction advice before score dimensions
