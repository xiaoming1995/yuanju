## ADDED Requirements

### Requirement: Year narratives SHALL explain concrete evidence in plain language

For a past-events year that has meaningful technical evidence, the system SHALL render a user-facing narrative that explains at least one concrete evidence trigger in language understandable to users who do not know bazi.

#### Scenario: Year has a branch clash evidence
- **WHEN** a year contains a meaningful signal whose evidence describes a branch clash
- **THEN** the year narrative SHALL explain that the year is more likely to involve change, movement, friction, re-negotiation, or adjustment
- **AND** the narrative SHALL connect the explanation to likely life areas when those areas are inferable from the signal theme or palace
- **AND** the narrative SHALL NOT only say a generic phrase such as "容易有变化" without explaining why

#### Scenario: Year has a shensha or special trigger evidence
- **WHEN** a year contains a meaningful signal whose evidence describes shensha, `夹拱`, `空亡`, `驿马`, `天乙贵人`, or another named trigger
- **THEN** the year narrative SHALL translate the trigger into everyday meaning
- **AND** the professional trigger name MAY remain visible when useful
- **AND** the narrative SHALL avoid requiring the user to understand the term before understanding the year

### Requirement: Year narratives SHALL use a readable explanation structure

For a meaningful past-events year, the system SHALL structure the visible narrative around yearly tendency, reason, likely life areas, and practical stance.

#### Scenario: Meaningful evidence supports a full narrative
- **WHEN** a year has enough meaningful signals and evidence to support a readable explanation
- **THEN** the narrative SHALL include an overall yearly tendency
- **AND** the narrative SHALL include a reason derived from the strongest evidence
- **AND** the narrative SHALL name likely affected life areas when inferable
- **AND** the narrative SHALL include a practical caution or suggested stance

#### Scenario: Evidence is weak or sparse
- **WHEN** a year has only weak or sparse evidence
- **THEN** the narrative SHALL remain conservative
- **AND** the narrative SHALL NOT invent specific events that are not supported by the available signals
- **AND** the narrative MAY fall back to a shorter evidence-based summary

### Requirement: Evidence-aligned fallback SHALL protect year-card accuracy

The system SHALL prefer an evidence-aligned deterministic narrative over generated year text when generated text is missing, too short, unsupported, or less specific than available evidence.

#### Scenario: Generated year text is missing
- **WHEN** a dayun summary does not provide a year narrative for a year that has meaningful evidence
- **THEN** the year card SHALL display the deterministic evidence-aligned narrative
- **AND** the card SHALL still expose the professional evidence detail separately

#### Scenario: Generated year text is too generic
- **WHEN** generated year text does not mention or explain any concrete trigger while deterministic evidence is available
- **THEN** the year card SHALL display or fall back to the deterministic evidence-aligned narrative
- **AND** generated text SHALL NOT make the year card less precise than the evidence summary

### Requirement: Professional evidence SHALL remain auditable

The system SHALL keep professional evidence available in the year card while making the main narrative readable for non-specialists.

#### Scenario: User reads a year card with technical evidence
- **WHEN** a year card renders an evidence-aligned narrative
- **THEN** the main visible content SHALL be understandable without opening the technical evidence detail
- **AND** the technical evidence detail SHALL remain available for users who want to audit the conclusion
- **AND** the technical evidence SHALL NOT be replaced by only natural-language paraphrase

### Requirement: Narrative terminology SHALL be conservative and non-absolute

The system SHALL describe past-events tendencies as likely manifestations rather than guaranteed events.

#### Scenario: Narrative explains a strong signal
- **WHEN** a year has strong evidence for a theme such as relationship change, movement, work pressure, money activity, learning, or support from others
- **THEN** the narrative SHALL use non-absolute wording such as "更容易体现为", "常见表现是", or equivalent phrasing
- **AND** the narrative SHALL NOT state unsupported certainties such as exact dates, guaranteed outcomes, or unverifiable events
