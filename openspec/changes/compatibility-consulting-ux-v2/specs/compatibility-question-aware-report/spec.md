## ADDED Requirements

### Requirement: Compatibility report generation uses relationship context
The compatibility AI prompt SHALL include relationship stage and primary user question when generating a compatibility report.

#### Scenario: Report is generated with context
- **WHEN** a user generates a compatibility report for a reading with relationship context
- **THEN** the prompt data includes the relationship stage and primary question
- **AND** the generated report prioritizes sections relevant to that question

#### Scenario: Report is generated without context
- **WHEN** a user generates a compatibility report for a legacy reading without context
- **THEN** the prompt uses a general relationship judgment mode
- **AND** report generation does not fail because context is missing

### Requirement: Report structure adapts to the primary question
The compatibility report SHALL use a question-aware structure so different relationship problems receive different emphasis.

#### Scenario: Reconciliation question
- **WHEN** `primary_question` is `reconciliation_potential`
- **THEN** the report covers whether reconciliation is advisable, what original problem may repeat, what signals must be verified, and what boundary conditions should stop the attempt

#### Scenario: Marriage suitability question
- **WHEN** `primary_question` is `marriage_suitability`
- **THEN** the report covers long-term stability, reality承接, conflict handling, family/responsibility/boundary risks, and questions to confirm before marriage decisions

#### Scenario: Continue investment question
- **WHEN** `primary_question` is `continue_investment`
- **THEN** the report covers whether to continue investing, what to observe next, how to pace commitment, and what behaviors to avoid

### Requirement: Report claims remain evidence-bound and conditional
The compatibility report SHALL keep major claims linked to structured evidence and avoid deterministic relationship promises.

#### Scenario: Report contains major claim
- **WHEN** the structured report returns a major decision claim
- **THEN** the claim includes supporting evidence keys when available
- **AND** the wording uses conditional relationship language rather than absolute fate language

#### Scenario: Evidence is insufficient
- **WHEN** the available structured evidence does not support a specific claim
- **THEN** the report does not invent unsupported bazi reasoning
- **AND** it uses an observation or uncertainty caveat instead
