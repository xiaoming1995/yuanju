# compatibility-professional-evidence-reporting Specification

## Purpose
TBD - created by archiving change compatibility-depth-signals. Update Purpose after archive.
## Requirements
### Requirement: Depth evidence in compatibility detail API
The system SHALL include depth-signal evidence in compatibility detail responses for readings analyzed with the enhanced engine.

#### Scenario: Detail response includes depth evidence
- **WHEN** a user opens a compatibility reading created or regenerated with the enhanced engine
- **THEN** the detail response SHALL include evidence items from the new depth signal families when applicable

#### Scenario: Legacy readings remain viewable
- **WHEN** a user opens a compatibility reading that lacks enhanced depth evidence
- **THEN** the detail response SHALL remain valid
- **AND** the frontend SHALL still render the existing compatibility result

### Requirement: Depth evidence in AI prompt data
The system SHALL include the enhanced compatibility evidence and score explanations in AI prompt data for compatibility reports.

#### Scenario: AI report generation receives structured evidence
- **WHEN** the system builds compatibility prompt data
- **THEN** it SHALL include grouped depth evidence and score explanations in structured form

#### Scenario: AI claims remain evidence-linked
- **WHEN** the AI prompt asks for relationship conclusions or advice
- **THEN** it SHALL instruct the model to ground major claims in provided evidence keys
- **AND** it SHALL prohibit exact-date or deterministic relationship predictions

### Requirement: Professional evidence grouping in frontend
The frontend SHALL render enhanced compatibility evidence in grouped professional sections while keeping the consultation summary as the primary reading path.

#### Scenario: User expands professional details
- **WHEN** a user opens the professional evidence area on a compatibility result
- **THEN** the page SHALL group evidence by signal family or relationship meaning
- **AND** it SHALL show plain-language detail before technical labels where both exist

#### Scenario: Evidence is absent for a group
- **WHEN** a specific depth signal family has no applicable evidence for the reading
- **THEN** the frontend SHALL omit that empty group rather than showing placeholder technical text

### Requirement: Report copy preserves uncertainty boundaries
The system SHALL present depth-signal interpretations as tendencies, risk factors, and support factors rather than absolute fate claims.

#### Scenario: Timing context is present
- **WHEN** timing evidence is included in a report or result page
- **THEN** the copy SHALL describe observation windows, pressure themes, or focus areas
- **AND** it SHALL NOT claim exact marriage, breakup, reconciliation, pregnancy, affair, or other deterministic event dates

#### Scenario: Evidence is mixed
- **WHEN** positive and negative depth evidence both exist
- **THEN** the report SHALL describe conditions and trade-offs rather than reducing the reading to a single absolute verdict

