## ADDED Requirements

### Requirement: Profile presents a continuation workbench
The profile page SHALL present a workbench section that guides users to continue their most relevant existing analysis or start a new one.

#### Scenario: Recent chart exists
- **WHEN** the profile response includes at least one recent chart
- **THEN** the workbench links to the most recent chart detail page

#### Scenario: No chart but recent compatibility exists
- **WHEN** the profile response has no recent chart and includes a recent compatibility reading
- **THEN** the workbench links to the most recent compatibility result page

#### Scenario: No recent records exist
- **WHEN** the profile response includes no recent chart or compatibility reading
- **THEN** the workbench links to the bazi input page

### Requirement: Profile stats guide users to existing destinations
The profile page SHALL render stats as navigable cards when an existing route supports the stat domain.

#### Scenario: Chart and compatibility stats
- **WHEN** profile stats render
- **THEN** chart count links to history and compatibility count links to compatibility history

### Requirement: Planned features are visibly inactive
The profile page SHALL clearly mark wallet and PDF template entries as planned features and MUST NOT link to payment or template routes.

#### Scenario: Planned features displayed
- **WHEN** wallet or PDF template feature cards render
- **THEN** each card shows a planned status badge and remains non-transactional
