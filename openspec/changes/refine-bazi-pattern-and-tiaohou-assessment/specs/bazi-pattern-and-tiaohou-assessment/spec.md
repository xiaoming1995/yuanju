## ADDED Requirements

### Requirement: Pattern formation and break require tiered evidence
The system SHALL distinguish visible, hidden, and rooted ten-god evidence when evaluating natal pattern formations and breaks. A hidden-only relevant ten god MUST NOT independently create a formation or break that changes the natal pattern score.

#### Scenario: Hidden Food God does not trigger Owl Seizing Food
- **WHEN** a natal chart is classified as 偏印格 and 食神 appears only in hidden stems
- **THEN** the system SHALL NOT add `枭神夺食` to the pattern breaks
- **AND THEN** the system SHALL NOT apply the corresponding pattern-score penalty

#### Scenario: Visible and rooted Food God triggers Owl Seizing Food
- **WHEN** a natal chart is classified as 偏印格 and 食神 is visible with effective branch root support
- **THEN** the system SHALL evaluate and record `枭神夺食` as a pattern break

### Requirement: Fuyi support is not a Pattern formation
The system SHALL record 印星与比劫的助身作用 through the Fuyi assessment and MUST NOT add `印比相扶` as a 偏印格 pattern formation solely because both groups occur in the natal chart.

#### Scenario: Visible 印比 support in a 偏印格
- **WHEN** a 偏印格 contains visible or hidden 印星与比劫 support
- **THEN** the Fuyi assessment SHALL retain the support evidence
- **AND THEN** the pattern formation list SHALL NOT contain `印比相扶`

### Requirement: Day-stem and thermal Tiaohou are separate assessments
The system SHALL provide separate day-stem Tiaohou and thermal Tiaohou assessments. Each assessment SHALL identify its required stems or elements, their visible/hidden availability, and its own status.

#### Scenario: Bing Fire in Xu month has dictionary support
- **WHEN** the system evaluates `乙亥·丙戌·丙子·甲午`
- **THEN** the day-stem Tiaohou assessment SHALL require `甲` and `壬`
- **AND THEN** it SHALL record `甲` as visible and `壬` as hidden
- **AND THEN** it SHALL not represent this result solely as a generic `调候不急`

#### Scenario: Hidden water partially remedies a thermal condition
- **WHEN** a seasonal thermal condition requires water and water exists only in hidden stems
- **THEN** the thermal Tiaohou assessment SHALL report partial hidden support
- **AND THEN** it SHALL not report the condition as fully resolved or absent solely because of that hidden support
