## ADDED Requirements

### Requirement: Compatibility entry is personality-first
The compatibility creation page SHALL frame the default reading around personality fit before presenting the two birth profile forms.

#### Scenario: User opens compatibility entry
- **WHEN** an authenticated user opens the compatibility creation page
- **THEN** the first major interaction area asks whether the user wants to understand personality fit, long-term fit, or repeated conflict
- **AND** the relationship stage selection appears before or alongside the question context
- **AND** the birth profile forms remain available on the same flow

#### Scenario: User changes the primary question
- **WHEN** the user selects a different primary question
- **THEN** the page updates the consultation preview to describe the personality-fit angle and validation path that reading will produce

### Requirement: Entry flow explains the reading outcome
The compatibility creation page SHALL preview the personality-fit output before the user starts the reading.

#### Scenario: Context preview is visible
- **WHEN** the user has selected a relationship stage and primary question
- **THEN** the page shows a concise preview of the expected output, such as relationship personality patterns, fit points, conflict points, and communication guidance

#### Scenario: Context is missing or legacy default
- **WHEN** context values are missing or set to general
- **THEN** the page presents a neutral general personality-fit preview without blocking reading creation

### Requirement: Entry flow remains mobile efficient
The compatibility creation page SHALL keep the consultation-first flow usable on mobile without hiding required birth data.

#### Scenario: User views entry flow on mobile
- **WHEN** the compatibility creation page is rendered on a narrow viewport
- **THEN** question controls, stage controls, and the active birth profile form stack in a readable order
- **AND** the primary start action remains reachable without overlapping the bottom navigation

### Requirement: Existing create-reading payload is preserved
The compatibility creation page SHALL continue sending the existing relationship context and birth profile data used by the current compatibility API.

#### Scenario: User starts a reading
- **WHEN** the user submits the compatibility creation form
- **THEN** the request includes both participant birth profiles plus the selected relationship stage and primary question
- **AND** no new required backend field is introduced by this UX change
