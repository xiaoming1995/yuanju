## ADDED Requirements

### Requirement: Generated interpretation preserves day-stem formation and other natal layers
The system SHALL include day-stem adjustment formation and its evidence in generated report and flow-year interpretation context, separately from seasonal thermal adjustment,扶抑用神, detected main-pattern quality, and 制化配合 evidence.

#### Scenario: Formed day-stem adjustment is supplied to an AI report
- **WHEN** report context is built for a chart whose day-stem adjustment formation is `formed`
- **THEN** the context SHALL state that `日干调候成格` provides a `高格基础`
- **AND** it SHALL include the required stems and the visible and hidden evidence
- **AND** it SHALL include the main-pattern quality and 制化配合 result as separate fields

#### Scenario: Other limitations coexist with a high foundation
- **WHEN** a chart has a high day-stem foundation together with thermal,扶抑, or critical-pattern limitations
- **THEN** generated context SHALL preserve each limitation
- **AND** it SHALL NOT present the high foundation as a guarantee of an unconditional final vehicle grade or life outcome
