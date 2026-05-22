## ADDED Requirements

### Requirement: Compatibility result includes personality validation plan
The compatibility result page SHALL present a short action plan that helps users validate the personality-fit judgment in real interaction.

#### Scenario: User opens a compatibility result
- **WHEN** a user opens a compatibility result detail
- **THEN** the page shows a validation plan after the personality-fit summary and decision summary, before dense score or professional evidence sections
- **AND** the plan includes short-term personality checks, medium-term interaction checks, and behaviors to avoid

#### Scenario: Structured AI report is absent
- **WHEN** the result has no AI report
- **THEN** the action plan still renders from deterministic compatibility assessment data
- **AND** the page makes clear that generating a deep report can refine the personality interpretation and plan

#### Scenario: Structured AI report exists
- **WHEN** the result has question-aware structured report data
- **THEN** the action plan can incorporate report advice while preserving the deterministic fallback order

### Requirement: Action plan uses verification language
The action plan SHALL frame guidance as observation, validation, and boundary-setting rather than deterministic fate claims.

#### Scenario: Plan text is rendered
- **WHEN** the action plan displays next steps
- **THEN** it uses conditional language such as observe communication rhythm, validate conflict repair, confirm emotional availability, avoid escalating too early, or set boundaries
- **AND** it does not predict exact breakup, marriage, reconciliation, pregnancy, affair, or event dates

### Requirement: Personality question is answered explicitly
The compatibility result page SHALL restate and directly answer the user's selected personality-fit or relationship-fit question in the result reading path.

#### Scenario: Result has primary question context
- **WHEN** the result page has a primary question value
- **THEN** the page shows the question label near the decision answer
- **AND** the personality-fit summary and action plan align their checks with that question

#### Scenario: Result lacks primary question context
- **WHEN** the result page lacks question context
- **THEN** the page falls back to general relationship judgment wording

### Requirement: Action plan links to supporting sections
The action plan SHALL provide a route to supporting stage-risk, score, or evidence sections when those sections are available.

#### Scenario: User inspects an action item
- **WHEN** an action item is based on stage risk or evidence data
- **THEN** the item exposes a visible affordance to view the related validation window, score dimension, or evidence area

#### Scenario: Supporting data is unavailable
- **WHEN** an action item has no supporting section available
- **THEN** the item remains readable and does not link to an empty destination
