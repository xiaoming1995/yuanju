## ADDED Requirements

### Requirement: Result overview renders exact stem-level Fuyi summary
When usable stem guidance is present, the Bazi result overview SHALL render concrete heavenly stems rather than only the aggregate Fuyi favorable and adverse element strings. The overview SHALL render a primary entry from `primary_favorable`, a complete `secondary_favorable` list, and a complete `adverse` list when those groups are available.

#### Scenario: A chart has shared primary and Fuyi-only stems
- **WHEN** `stem_guidance` contains a primary favorable stem, Fuyi-only favorable stems, and adverse stems
- **THEN** the overview SHALL show the first primary item as `首取：<stem><element> · <ten_god>` with its day-stem-adjustment and Fuyi source
- **AND** it SHALL list every secondary and adverse item by concrete stem and element
- **AND** it SHALL not present only aggregate strings such as `金土火` or `水木` in place of those lists

#### Scenario: 1996 chart explains its concrete Fuyi direction
- **WHEN** the 1996-02-08 20:00 chart has `丙火` as primary, `庚金、辛金、戊土、己土、丁火` as secondary, and `壬水、甲木、乙木` as adverse
- **THEN** the overview SHALL identify `丙火 · 伤官` as the first choice
- **AND** it SHALL display the listed secondary stems under `扶抑可用`
- **AND** it SHALL display the listed adverse stems under `扶抑慎用`

### Requirement: Overview preserves recommendation boundaries
The result overview SHALL reserve `首取` for a primary favorable item jointly supported by day-stem adjustment and Fuyi. It SHALL label Fuyi-only favorable stems as `扶抑可用`, label adverse stems as `扶抑慎用`, and SHALL NOT render `conditioning_only` stems as general future favorable stems.

#### Scenario: A chart has no shared primary favorable stem
- **WHEN** `primary_favorable` is empty while secondary favorable or adverse guidance exists
- **THEN** the overview SHALL omit `首取`
- **AND** it SHALL render the available groups without inventing a first-choice heavenly stem

#### Scenario: A chart has adjustment-only conflict information
- **WHEN** a stem appears only in `conditioning_only` because day-stem adjustment conflicts with Fuyi
- **THEN** the overview SHALL not include that stem in `首取` or `扶抑可用`
- **AND** the professional detail surface SHALL retain its existing conflict explanation

### Requirement: Overview remains backward compatible and responsive
The result overview SHALL retain its element-level Fuyi badges when a response has no usable `stem_guidance`. Exact stem-level summary rows SHALL remain readable without overlap or truncation at supported desktop and mobile widths.

#### Scenario: A legacy chart lacks stem guidance
- **WHEN** a stored Bazi result has no `natal_assessment.stem_guidance`
- **THEN** the overview SHALL render the existing aggregate Fuyi favorable and adverse element badges
- **AND** it SHALL not display an empty stem-level summary

#### Scenario: Exact stem lists wrap on a narrow screen
- **WHEN** a result overview is viewed on a narrow mobile viewport
- **THEN** its primary, usable, and adverse summary text SHALL wrap within the overview container
- **AND** the pillar heading, reading-mode controls, and summary content SHALL remain unobscured

### Requirement: Exact summary is presentation-only
The system SHALL treat the exact stem summary as a presentation of existing `stem_guidance`. It SHALL NOT alter any natal-assessment score, grade, vehicle category, Dayun road score, or annual-flow score.

#### Scenario: A known chart is recalculated after the summary change
- **WHEN** a Bazi result is rendered with exact stem-level summary enabled
- **THEN** its pre-existing natal grade, vehicle category, Dayun-road result, and annual-flow calculation SHALL remain unchanged
