## ADDED Requirements

### Requirement: Compatibility result page presents decision-first consultation
The compatibility result page SHALL present the relationship decision and next action before score tables and professional bazi details.

#### Scenario: User opens a compatibility result
- **WHEN** a user opens a compatibility result detail
- **THEN** the first major content area shows the relationship context, decision headline, core contradiction, and next recommended actions
- **AND** score details appear after the decision-oriented content

#### Scenario: AI report is not generated
- **WHEN** no AI compatibility report exists for the reading
- **THEN** the result page still renders deterministic consulting guidance from saved compatibility assessment data
- **AND** the page offers a single primary action to generate the full report when relevant

### Requirement: Compatibility scores are explained as user questions
The result page SHALL translate the four compatibility dimensions into plain relationship questions while preserving the underlying score values.

#### Scenario: Scores are rendered
- **WHEN** the result page renders dimension scores
- **THEN** `attraction` is presented as whether the pair is mutually attracted
- **AND** `stability` is presented as whether the relationship can stay stable long term
- **AND** `communication` is presented as whether conflicts can be repaired
- **AND** `practicality` is presented as whether reality conditions can land

### Requirement: Stage risks are shown as relationship tasks
The result page SHALL present duration windows as stage tasks and risks rather than fate-like timing predictions.

#### Scenario: Stage risks are rendered
- **WHEN** the result page renders the three duration windows
- **THEN** each window includes the stage purpose, main risk, trigger, and suggested action
- **AND** the UI avoids exact breakup or marriage timing language

### Requirement: Professional evidence remains expandable
The result page SHALL keep professional chart evidence available without making it the default first reading path.

#### Scenario: User wants to inspect evidence
- **WHEN** a user expands professional evidence or a claim's evidence link
- **THEN** the page shows the relevant bazi evidence cards and reasoning
- **AND** the evidence remains connected to the claim through stable evidence keys when available
