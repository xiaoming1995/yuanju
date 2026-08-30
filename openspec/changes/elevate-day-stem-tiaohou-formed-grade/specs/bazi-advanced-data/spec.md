## ADDED Requirements

### Requirement: Versioned day-stem adjustment formation assessment
The backend SHALL expose a versioned day-stem adjustment assessment containing availability status, formation state, foundation tier, foundation-score contribution, required stems, visible-stem evidence, and hidden-stem evidence. Existing day-stem adjustment status and score fields MUST remain available for backward compatibility.

#### Scenario: Required stems span heaven and earth
- **WHEN** every stem required by the day-stem adjustment dictionary is present across the visible heavenly stems and the earthly-branch hidden stems, with at least one evidence item in each location
- **THEN** the assessment SHALL set `formation` to `formed`
- **AND** it SHALL set `foundation_tier` to `high`
- **AND** it SHALL identify the conclusion as `日干调候成格`
- **AND** it SHALL expose the complete visible and hidden evidence used for that conclusion

#### Scenario: Required stems are only partially covered
- **WHEN** one or more required day-stem adjustment stems are missing, or all matched evidence occurs in only one of the two locations
- **THEN** the assessment SHALL NOT set `formation` to `formed`
- **AND** it SHALL retain a conservative partial or unformed availability and foundation result

#### Scenario: Extra stems cannot satisfy missing required stem evidence
- **WHEN** the chart contains a number of matched or unrelated stems equal to the number of required stems but does not cover every specific required stem
- **THEN** the assessment SHALL treat the missing required stem as unresolved
- **AND** it SHALL NOT classify the day-stem adjustment as `formed`

### Requirement: High day-stem foundation remains distinct from main-pattern quality
The natal assessment SHALL expose the day-stem adjustment foundation independently from the detected main pattern's root and 制化 quality.

#### Scenario: Formed day-stem adjustment with partial main pattern
- **WHEN** the day-stem adjustment is `formed` and the detected main pattern lacks sufficient root or 制化 evidence for generic pattern formation
- **THEN** the assessment SHALL retain `high` as the day-stem foundation tier
- **AND** it SHALL retain the main pattern's own quality result independently
- **AND** it SHALL NOT describe the day-stem adjustment as unformed solely because of the main pattern's quality

#### Scenario: Main-pattern transformation is present
- **WHEN** the detected main pattern satisfies an existing 制化配合 rule
- **THEN** the assessment SHALL preserve that transformation evidence independently of the day-stem adjustment formation result

### Requirement: Vehicle-grade calculation accounts for high day-stem foundation transparently
The natal vehicle-grade calculation SHALL include the explicit day-stem foundation contribution in its score and evidence trail while retaining urgent thermal,扶抑, and critical-break constraints.

#### Scenario: Formed day-stem adjustment contributes to vehicle assessment
- **WHEN** a natal assessment has `formation` equal to `formed`
- **THEN** the vehicle evidence list SHALL include an auditable `日干调候成格` entry with the foundation contribution and evidence
- **AND** the final score SHALL use that contribution rather than only a generic availability score

#### Scenario: Other hard constraints remain effective
- **WHEN** a natal assessment has a formed high day-stem foundation and an unresolved urgent thermal condition, insufficient 扶抑 support, or a critical main-pattern break
- **THEN** the final vehicle tier SHALL continue to apply the relevant existing limitation
- **AND** the output SHALL state the high foundation and the limitation as separate conclusions
