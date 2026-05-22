## ADDED Requirements

### Requirement: Absent deep report state is compact and actionable
The compatibility result page SHALL render a compact, explanatory, actionable state when no deep report has been generated.

#### Scenario: No deep report exists
- **WHEN** the compatibility result has no latest report
- **THEN** the deep report section explains that the structured compatibility reading is already available and that AI deep reading is optional
- **AND** the section exposes the single report-generation action in the local context of the deep report section

#### Scenario: No deep report section is shown below completed result data
- **WHEN** the user scrolls to the deep report section after reading scores and evidence
- **THEN** the section does not appear as a large empty content card

### Requirement: Deep report generation state is clear
The compatibility result page SHALL communicate loading and error states for deep report generation near the deep report module.

#### Scenario: Report generation is loading
- **WHEN** the user starts generating a deep report
- **THEN** the deep report section shows a loading state and disables duplicate generation actions

#### Scenario: Report generation fails
- **WHEN** report generation returns an error
- **THEN** the deep report section shows the error near the generation action and allows retry

### Requirement: Generated deep reports are readable
The compatibility result page SHALL display generated deep report content with readable hierarchy.

#### Scenario: Structured report exists
- **WHEN** the latest report includes structured content
- **THEN** the deep report section renders question focus, summary, dimensions, risks, and advice as distinct readable blocks

#### Scenario: Only raw report exists
- **WHEN** the latest report has raw content but no structured content
- **THEN** the deep report section renders the raw content in a readable fallback block

### Requirement: Deep report generation action remains singular
The compatibility result page SHALL avoid multiple competing visible actions for generating the same deep report.

#### Scenario: No report exists
- **WHEN** the page renders without a latest report
- **THEN** exactly one visible primary action generates the deep report
