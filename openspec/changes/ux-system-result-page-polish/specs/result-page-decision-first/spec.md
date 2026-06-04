## ADDED Requirements

### Requirement: Result page decision-first hero

The result page SHALL present a decision-first hero area before professional detail sections.

#### Scenario: First screen shows chart identity and conclusion
- **WHEN** a user opens the result page for a calculated or saved chart
- **THEN** the first screen SHALL show the chart identity summary, four-pillar overview, and a plain-language core conclusion before detailed professional modules.

#### Scenario: First screen shows primary action
- **WHEN** a user opens the result page
- **THEN** the first screen SHALL show the primary action to generate or view AI interpretation and secondary actions for past events and export when available.

#### Scenario: Missing AI report does not block summary
- **WHEN** no AI report exists for the chart
- **THEN** the result page SHALL still render a deterministic chart summary and a clear primary action to generate the AI interpretation.

### Requirement: Result page segmented navigation

The result page SHALL group detailed content into stable sections with a segmented navigation affordance.

#### Scenario: Result sections are available
- **WHEN** a user views the result page
- **THEN** the page SHALL expose sections for overview, chart details, useful-god analysis, major luck, and AI interpretation.

#### Scenario: Navigation targets are stable
- **WHEN** a user activates a result page segment
- **THEN** the page SHALL navigate to the matching section without losing chart state or requiring a new calculation.

#### Scenario: Mobile segment navigation remains usable
- **WHEN** the result page is viewed on a narrow mobile viewport
- **THEN** the segmented navigation SHALL remain horizontally usable without forcing page-level horizontal overflow.

### Requirement: Mobile primary action remains reachable

The result page SHALL keep the primary action reachable on mobile without obscuring content.

#### Scenario: Mobile fixed action
- **WHEN** a user views the result page on a mobile viewport
- **THEN** the primary AI interpretation action SHALL be reachable from a bottom action area or equivalent persistent control.

#### Scenario: Bottom navigation is not obscured
- **WHEN** the mobile primary action is visible
- **THEN** it SHALL account for existing bottom navigation and safe-area spacing so page content and navigation remain operable.

### Requirement: Related result actions are consistent

The result page, history page, and past-events page SHALL use consistent action labels and routing for AI interpretation, past events, and export.

#### Scenario: History card action labels
- **WHEN** a user views a saved chart in history
- **THEN** actions leading to the result page, past events, or export SHALL use labels consistent with the result page actions.

#### Scenario: Past events empty state
- **WHEN** a user opens past events without usable chart context
- **THEN** the page SHALL show an inline empty state with a route back to calculation or history instead of leaving the user at a dead end.
