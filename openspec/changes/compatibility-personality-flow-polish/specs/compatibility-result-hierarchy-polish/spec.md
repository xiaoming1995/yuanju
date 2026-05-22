## ADDED Requirements

### Requirement: Result page provides a compact reading map
The compatibility result page SHALL provide a compact navigation map for the main reading sections.

#### Scenario: User opens a result
- **WHEN** a user opens a compatibility result detail
- **THEN** the page shows quick links for personality fit, conflict/action validation, scores or evidence, and professional details
- **AND** each link targets an existing section that is present on the page

#### Scenario: User views on mobile
- **WHEN** the result page is rendered on a narrow viewport
- **THEN** the reading map remains usable without hiding content or overlapping bottom navigation

### Requirement: Validation content is not duplicated
The compatibility result page SHALL present personality validation and stage-risk validation as one coherent reading path.

#### Scenario: User reads validation guidance
- **WHEN** the page shows 7-day and 30-day validation guidance
- **THEN** related stage-risk details are grouped under, linked from, or visually attached to the validation plan
- **AND** the page avoids presenting two adjacent modules that appear to answer the same "what to validate next" question

### Requirement: Personality match types are self-explanatory
The compatibility result page SHALL explain the displayed personality match type with concise user-facing copy.

#### Scenario: Match type appears
- **WHEN** a personality match type such as "高吸引高消耗型" is displayed
- **THEN** the page also shows a short explanation of what that type means in everyday interaction
- **AND** the explanation avoids deterministic fate claims

### Requirement: Personality section uses lower-density layout
The compatibility result page SHALL reduce nested-card density in personality and validation areas.

#### Scenario: User scans personality fit
- **WHEN** the user scans fit points, conflict points, and communication advice
- **THEN** the section uses stable grouped rows or panels rather than excessive nested cards
- **AND** text remains readable across desktop and mobile viewports
