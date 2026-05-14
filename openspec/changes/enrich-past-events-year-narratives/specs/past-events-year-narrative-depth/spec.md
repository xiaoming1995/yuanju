## ADDED Requirements

### Requirement: Medium-detail yearly narratives
The system SHALL generate medium-detail narrative text for past-events year cards when the year has meaningful rule-based signals.

#### Scenario: Signal-bearing year has richer body text
- **WHEN** a past-events year contains one or more meaningful event signals
- **THEN** the generated `narrative` SHALL include yearly tone, concrete life-domain manifestation, and a practical stance rather than only a short reminder sentence

#### Scenario: Weak-signal year remains concise but useful
- **WHEN** a past-events year has no meaningful event signals
- **THEN** the generated `narrative` SHALL remain concise while still explaining the year as comparatively calm or weakly triggered

### Requirement: Narrative remains deterministic and token-free
The system SHALL generate enhanced year-card narratives from local rule-based data without adding per-year LLM calls.

#### Scenario: Generate past-events years
- **WHEN** the past-events years endpoint generates year cards
- **THEN** yearly narratives SHALL be produced from existing bazi signals, ten-god power, and template rules without invoking an LLM for each year

### Requirement: Domain-specific details
The system SHALL translate yearly signals into domain-specific plain-language details.

#### Scenario: Study-age year
- **WHEN** the native age is below the study-age cutoff and the year has study, peer, rule, or resource signals
- **THEN** the narrative SHALL use school, teachers, peers, family resources, routine, or examination wording instead of adult career or romance defaults

#### Scenario: Adult domain year
- **WHEN** the native is an adult and the year has career, money, relationship, health, movement, or support signals
- **THEN** the narrative SHALL include the relevant real-life domain rather than generic change wording alone

### Requirement: Hard event evidence priority
The system SHALL keep hard event evidence prominent when a year contains strong clash, punishment, void, fuyin, fanyin, heavy formation, or yongshen/jishen hit signals.

#### Scenario: Hard event plus ten-god force
- **WHEN** a year contains both hard event evidence and a ten-god power profile
- **THEN** the narrative SHALL lead with the hard event's practical meaning and use ten-god power only as context if it adds useful explanation

#### Scenario: Technical evidence remains expandable
- **WHEN** a year has technical evidence used to build the narrative
- **THEN** the default card body SHALL keep wording readable and the detailed technical basis SHALL remain available through evidence summary

### Requirement: Ten-god force integration
The system SHALL integrate ten-god force into yearly narratives in plain language when it clarifies the year.

#### Scenario: Useful ten-god context
- **WHEN** a year has a ten-god power profile that adds a distinct driving force to the selected event signals
- **THEN** the narrative SHALL explain that force using plain-language concepts such as money/resources, rules/responsibility, study/support, expression/output, or peers/competition

#### Scenario: Avoid repeated standalone force sentence
- **WHEN** adjacent years share the same broad ten-god force but have different concrete event signals
- **THEN** the generated narratives SHALL not rely on the exact same standalone ten-god sentence as the main differentiator

### Requirement: Timeline readability
The system SHALL keep longer year-card narratives readable in the past-events timeline.

#### Scenario: Desktop and mobile layout
- **WHEN** enhanced narratives are displayed on desktop or mobile widths
- **THEN** the year card layout SHALL preserve readable line height, spacing, and wrapping without overlapping badges, force rows, evidence toggles, or neighboring cards
