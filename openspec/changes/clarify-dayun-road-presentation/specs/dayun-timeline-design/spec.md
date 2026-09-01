## MODIFIED Requirements

### Requirement: Dayun summary strip appears below Liunian panel
The system SHALL render a Dayun summary strip below the Liunian cards with a concise, layered explanation of the selected Dayun. It SHALL make the ten-year composite road condition visually primary and keep phase prompts and technical tags secondary.

#### Scenario: Summary strip renders on desktop
- **WHEN** the Liunian panel has rendered
- **THEN** a `大运总览` strip appears below it with a short plain-language conclusion and compact tags such as ten-god main qi, five-element main qi, and trend keywords
- **AND** when road data is available, it SHALL label `road_label` as `十年综合路况`
- **AND** it SHALL render front-five and back-five information as Jin Bu Huan phase prompts rather than an equivalent road-grade pair.

#### Scenario: Summary strip remains compact
- **WHEN** summary content is long
- **THEN** the strip preserves a concise wrapped layout without pushing the timeline into a verbose report section
- **AND** detailed phase evidence and professional terminology remain behind the existing secondary disclosure paths.

#### Scenario: Timeline card content matches mockup density
- **WHEN** a Dayun card renders
- **THEN** it displays period index, age range, ganzhi, gan/zhi Ten-God labels, composite road badge, and Gregorian year range
- **AND** it SHALL NOT render front-five and back-five phase prompts inside every compact timeline card.
