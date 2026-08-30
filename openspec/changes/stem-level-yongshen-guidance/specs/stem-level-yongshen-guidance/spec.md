## ADDED Requirements

### Requirement: Bazi result exposes additive stem-level Fuyi guidance
The system SHALL expose an optional `natal_assessment.stem_guidance` object while preserving existing Fuyi element-level `yongshen` and `jishen` fields. Each listed stem SHALL include its stem, five element, Ten God relative to the day master, source layers, and an explanatory detail.

#### Scenario: Existing clients consume element-level Fuyi fields
- **WHEN** an API client reads a calculated Bazi result with stem guidance
- **THEN** `natal_assessment.fuyi.yongshen` and `natal_assessment.fuyi.jishen` SHALL retain their existing complete element-level meanings
- **AND** the client SHALL receive stem guidance as an additive field

#### Scenario: Older response has no stem guidance
- **WHEN** a frontend receives a stored or older Bazi result without `stem_guidance`
- **THEN** it SHALL continue to render the existing element-level Fuyi conclusion without error

### Requirement: Shared day-stem adjustment and Fuyi stems are primary favorable stems
The system SHALL classify an ordered day-stem adjustment required stem as `primary_favorable` only when that stem's five element is included in the global Fuyi favorable-element conclusion. The item's source layers SHALL identify both `日干调候` and `扶抑`.

#### Scenario: A required stem is also Fuyi favorable
- **WHEN** a chart's day-stem adjustment requires `壬` and global Fuyi selects water as favorable
- **THEN** `壬水` SHALL appear in `primary_favorable`
- **AND** it SHALL include its Ten God and both source layers

#### Scenario: No exact cross-layer stem exists
- **WHEN** no day-stem adjustment required stem belongs to a Fuyi favorable element
- **THEN** `primary_favorable` SHALL be empty
- **AND** the system SHALL NOT choose an arbitrary yin or yang stem as a primary favorable stem

### Requirement: Fuyi-only and conflicting stems remain distinguishable
The system SHALL list both heavenly stems for each Fuyi favorable element as `secondary_favorable` unless already primary. A day-stem adjustment required stem that conflicts with the Fuyi adverse-element conclusion SHALL appear only in `conditioning_only` and SHALL state that it is not a general future Fuyi favorable stem.

#### Scenario: A day-stem adjustment requirement conflicts with Fuyi
- **WHEN** a required adjustment stem belongs to a Fuyi adverse element
- **THEN** the stem SHALL appear in `conditioning_only` with an explicit conflict explanation
- **AND** it SHALL NOT appear in `primary_favorable`, `secondary_favorable`, or the visible adverse list

#### Scenario: Fuyi-only favorable stems are listed
- **WHEN** Fuyi selects water and metal as favorable and `壬` is primary
- **THEN** `癸水`、`庚金`、and `辛金` SHALL appear in `secondary_favorable`
- **AND** `壬水` SHALL not be duplicated in that group

### Requirement: Result page presents concise stem-level guidance
The result page SHALL preserve the compact Fuyi element badges and SHALL expose the stem-level groups through a concise summary and progressive detail surface. The display SHALL identify primary favorable, secondary favorable, adjustment-only, and adverse stems using their source and Ten God labels.

#### Scenario: Primary favorable stem is available
- **WHEN** the calculated result has at least one primary favorable stem
- **THEN** the Fuyi summary SHALL identify the first primary stem as the heavenly-stem priority
- **AND** the detail surface SHALL display every available guidance group without presenting adjustment-only stems as general favorable stems

#### Scenario: Fuyi conclusion is neutral
- **WHEN** Fuyi provides no favorable or adverse element conclusion
- **THEN** the result page SHALL not invent favorable or adverse heavenly stems
- **AND** it SHALL retain any available day-stem adjustment reference with its source context

### Requirement: Stem guidance does not alter established scoring
The system SHALL treat stem-level guidance as explanatory data in this change. Existing natal grade, vehicle, Dayun-road, and annual-flow scoring SHALL continue to use their established inputs and thresholds.

#### Scenario: Existing Dayun fixture is recalculated
- **WHEN** a previously verified chart is recalculated with stem guidance enabled
- **THEN** its Dayun road type, score, aggregate evidence, and phase evidence SHALL remain unchanged unless independently modified by an existing scoring rule
