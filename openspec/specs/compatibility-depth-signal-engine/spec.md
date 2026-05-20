# compatibility-depth-signal-engine Specification

## Purpose
TBD - created by archiving change compatibility-depth-signals. Update Purpose after archive.
## Requirements
### Requirement: Ten-god relationship signals
The system SHALL compute directional ten-god relationship signals for compatibility readings when both participants have sufficient chart data.

#### Scenario: Directional ten-god meaning is generated
- **WHEN** the compatibility engine analyzes two valid bazi charts
- **THEN** it SHALL produce evidence describing how each participant's day master tends to experience the other participant's relevant elements or stems
- **AND** directional evidence SHALL identify whether the meaning applies from self to partner, partner to self, or both

#### Scenario: Missing ten-god inputs do not block analysis
- **WHEN** ten-god inputs cannot be derived from available chart data
- **THEN** the compatibility engine SHALL skip ten-god relationship evidence
- **AND** it SHALL still return the remaining compatibility analysis successfully

### Requirement: Favorable-element support signals
The system SHALL evaluate whether each participant tends to support, drain, or aggravate the other participant's structural element balance using conservative evidence language.

#### Scenario: Support tendency is identified
- **WHEN** one participant's chart strongly supplies an element the other participant lacks or benefits from according to available chart structure
- **THEN** the compatibility engine SHALL add positive support evidence with the source `favorable_element_support`

#### Scenario: Pressure tendency is identified
- **WHEN** one participant's chart strongly amplifies an element imbalance or pressure pattern in the other participant's chart
- **THEN** the compatibility engine SHALL add negative or mixed evidence with the source `favorable_element_support`

#### Scenario: Full yongshen precision is unavailable
- **WHEN** the system lacks a verified yongshen model for either participant
- **THEN** favorable-element evidence SHALL use tendency-based wording
- **AND** it SHALL NOT claim definitive yongshen or jishen conclusions

### Requirement: Expanded gan-zhi interaction signals
The system SHALL evaluate relationship-relevant heavenly-stem and earthly-branch interactions across all four pillars.

#### Scenario: Stem combinations are detected
- **WHEN** heavenly stems across the two charts form supported combination patterns
- **THEN** the compatibility engine SHALL add attraction, communication, or stability evidence according to the interaction meaning and affected pillars

#### Scenario: Branch interactions are detected
- **WHEN** earthly branches across the two charts form supported combination, meeting, clash, punishment, harm, or break patterns
- **THEN** the compatibility engine SHALL add positive, negative, or mixed evidence according to the interaction meaning and affected pillars

#### Scenario: Relationship-relevant pillars receive priority
- **WHEN** the same interaction type appears in multiple pillar positions
- **THEN** day-pillar and spouse-palace interactions SHALL carry higher relationship weight than year-pillar background interactions

### Requirement: Relationship-pattern synthesis signals
The system SHALL synthesize low-count pattern signals for communication style, conflict trigger, attachment/security tendency, reality pressure, and repairability from the deterministic evidence set.

#### Scenario: Pattern signals summarize repeated evidence
- **WHEN** multiple evidence items indicate the same relationship pattern
- **THEN** the compatibility engine SHALL add a summarized relationship-pattern evidence item
- **AND** that item SHALL reference or be traceable to the underlying evidence family

#### Scenario: Pattern signals remain bounded
- **WHEN** relationship-pattern evidence is generated
- **THEN** it SHALL NOT dominate score movement over direct chart interaction evidence

