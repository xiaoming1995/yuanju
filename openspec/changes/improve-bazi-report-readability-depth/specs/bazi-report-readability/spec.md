## ADDED Requirements

### Requirement: Detailed Report Default View
The frontend SHALL default structured bazi reports to the detailed interpretation view while retaining a control that allows the user to switch to the concise view.

#### Scenario: New structured report opens in detailed mode
- **WHEN** a user views a bazi result whose report contains `content_structured`
- **THEN** the report SHALL initially render the detailed chapter content
- **THEN** the concise chapter content SHALL remain accessible through the report mode switcher

#### Scenario: Legacy report fallback remains available
- **WHEN** a user views a report without `content_structured`
- **THEN** the frontend SHALL render the legacy `content` text without showing a broken or irrelevant mode switcher

### Requirement: Long-form Report Typography
The frontend SHALL style bazi report text for comfortable long-form Chinese reading.

#### Scenario: Report body is readable
- **WHEN** the report body is rendered on the result page
- **THEN** chapter body text SHALL use a readable long-form size and line height suitable for multi-paragraph Chinese content
- **THEN** the report title, summary, and mode switcher SHALL remain visually subordinate to the main report content

#### Scenario: Mobile report remains readable
- **WHEN** the report is viewed on a mobile viewport
- **THEN** body text SHALL remain readable without horizontal overflow or clipped content

### Requirement: User-friendly Mode Labels
The frontend SHALL use report mode labels that communicate the difference between short conclusions and full interpretation.

#### Scenario: User can understand report mode choices
- **WHEN** the mode switcher is displayed
- **THEN** the labels SHALL clearly distinguish concise reading from full detailed reading
- **THEN** the labels SHALL avoid making the detailed mode sound useful only to professional users
