## ADDED Requirements

### Requirement: Evidence-backed score contribution
The system SHALL only adjust compatibility dimension scores through generated evidence items.

#### Scenario: Score movement has evidence
- **WHEN** a compatibility score differs from its baseline
- **THEN** the returned analysis SHALL include one or more evidence items whose dimensions and weights explain that score movement

#### Scenario: No hidden score adjustment
- **WHEN** the compatibility engine applies a scoring rule
- **THEN** it SHALL create or update an evidence item for the affected dimension

### Requirement: Bounded source contribution
The system SHALL bound the impact of each signal source on final compatibility scores to prevent one evidence family from overwhelming the reading.

#### Scenario: Repeated interactions are capped
- **WHEN** multiple evidence items from the same source affect the same dimension
- **THEN** the scoring layer SHALL apply a source contribution cap before returning the final dimension score

#### Scenario: Scores remain within allowed range
- **WHEN** all evidence contributions have been applied
- **THEN** each compatibility dimension score SHALL remain within the existing public score range

### Requirement: Stable evidence taxonomy
The system SHALL use stable evidence source identifiers for new signal families so reports, UI grouping, and tests can rely on them.

#### Scenario: New signal sources are identifiable
- **WHEN** the engine returns depth evidence
- **THEN** each item SHALL include one of the stable source identifiers for ten-god interaction, favorable-element support, gan-zhi interaction, relationship pattern, or timing context

#### Scenario: Existing evidence consumers remain compatible
- **WHEN** a frontend or API client reads compatibility evidence
- **THEN** existing fields such as title, detail, dimension, polarity, source, and weight SHALL remain available

### Requirement: Score explanation summary
The system SHALL provide a concise score explanation that identifies the strongest positive and negative forces behind each compatibility dimension.

#### Scenario: Dimension has mixed evidence
- **WHEN** a dimension contains both positive and negative evidence
- **THEN** the explanation SHALL surface both the strongest support factor and the strongest pressure factor

#### Scenario: Dimension has limited evidence
- **WHEN** a dimension has too little evidence for a confident explanation
- **THEN** the explanation SHALL use neutral language and avoid overstating certainty
