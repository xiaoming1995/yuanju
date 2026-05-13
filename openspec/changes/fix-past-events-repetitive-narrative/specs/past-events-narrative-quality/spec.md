## ADDED Requirements

### Requirement: Yearly narratives avoid repetitive generic openings
The system SHALL avoid using the same generic opening sentence for consecutive past-event years when those years contain materially different non-generic signals.

#### Scenario: Adjacent years share broad change but differ in specific signals
- **WHEN** adjacent yearly signal sets both contain `综合变动`
- **AND** one year also contains a school signal while another contains a health or relationship signal
- **THEN** their default `narrative` opening sentence SHALL differ according to the specific signal focus

#### Scenario: Only generic change signal exists
- **WHEN** a yearly signal set contains only broad `综合变动` and no more specific user-life theme
- **THEN** the default `narrative` SHALL use change-focused wording
- **AND** the wording SHALL still be plain language rather than raw technical evidence

### Requirement: Specific themes outrank ordinary broad change
The system SHALL prioritize specific user-life themes over ordinary broad `综合变动` when generating the default past-event yearly `narrative`.

#### Scenario: School signal and ordinary change coexist for a child-age year
- **WHEN** a year has `age < 18`
- **AND** its signals include both ordinary `综合变动` and a school-related type such as `学业_压力`, `学业_贵人`, `学业_才艺`, or `学业_竞争`
- **THEN** the default `narrative` SHALL focus on learning, teachers, exams, rules, or school environment before generic change language

#### Scenario: Health signal and ordinary change coexist
- **WHEN** a year has both ordinary `综合变动` and `健康`
- **THEN** the default `narrative` SHALL include a health, sleep, energy, safety, or recovery focus
- **AND** ordinary change SHALL NOT be the only visible yearly interpretation

### Requirement: Strong change remains visible
The system SHALL still allow strong change signals to become the dominant narrative theme when the evidence indicates a major-change year.

#### Scenario: Fuyin or fanyin exists
- **WHEN** a yearly signal set includes `伏吟` or `反吟`
- **THEN** the default `narrative` SHALL mention repeated issues, old matters, or stronger-than-usual change in plain language

#### Scenario: Dayun-liunian double hit exists
- **WHEN** a yearly signal evidence contains "大运流年双重命中", "力度倍增", or "重大事件高发"
- **THEN** the default `narrative` SHALL treat the year as a strong-change year

### Requirement: Child-age years use child-appropriate domains
For `age < 18`, the system SHALL prefer child-age domains in default yearly narratives and avoid adult-first wording unless no child-age signal exists.

#### Scenario: Child-age year has school and peer signals
- **WHEN** a year has `age < 18`
- **AND** its signals include school or personality types
- **THEN** the default `narrative` SHALL refer to learning, school, teachers, classmates, family communication, emotional development, or routine
- **AND** it SHALL NOT lead with adult-first career, finance, or romance framing

### Requirement: Professional evidence remains preserved
The system SHALL continue preserving professional technical evidence in `evidence_summary` even when the default `narrative` is plain-language and non-repetitive.

#### Scenario: Narrative removes raw technical terms
- **WHEN** `RenderYearNarrative` converts signals into plain-language output
- **THEN** raw terms such as `流年地支`, `月柱`, `官杀`, `伏吟`, `空亡`, or `财星` SHALL NOT appear in the default `narrative`
- **AND** the relevant technical evidence SHALL remain available from `RenderEvidenceSummary`
