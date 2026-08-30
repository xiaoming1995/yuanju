## ADDED Requirements

### Requirement: Evaluate Fuyi strength from the complete natal chart
The system SHALL derive the regular Fuyi strength level from all four pillars, including visible stems, ordered hidden stems, month-command weighting, day-master roots, and Jianlu/Yangren structural support. It SHALL expose the resulting level, score, favorable elements, adverse elements, and human-readable evidence to all internal consumers.

#### Scenario: Strongly supported Bing-fire natal chart
- **WHEN** the system evaluates `乙亥·丙戌·丙子·甲午` with `丙` as day master
- **THEN** it SHALL classify the regular Fuyi strength as `vstrong`
- **AND THEN** its favorable elements SHALL identify water as primary and metal as supporting, without labelling wood-fire as favorable Fuyi elements

#### Scenario: Hidden stems retain positional weight
- **WHEN** the system scores an earthly branch with main, middle, and residual hidden stems
- **THEN** it SHALL give the main hidden stem more weight than middle and residual hidden stems
- **AND THEN** it SHALL not decide strength by counting hidden-ten-god names alone

### Requirement: Keep Tiaohou and Fuyi conclusions distinct
The system SHALL retain Tiaohou and Fuyi conclusions as separate semantic outputs. A Tiaohou urgency MAY control priority in interpretation, but it MUST NOT overwrite the Fuyi favorable-element result.

#### Scenario: Tiaohou differs from Fuyi
- **WHEN** a chart has Tiaohou guidance different from its global Fuyi favorable elements
- **THEN** the result SHALL preserve both values
- **AND THEN** Fuyi consumers SHALL receive the global Fuyi favorable elements
