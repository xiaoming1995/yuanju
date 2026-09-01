## ADDED Requirements

### Requirement: Each available Dayun phase exposes a plain-language road and theme
The system SHALL present each available front-five or back-five Dayun phase with a phase road label and a concise user-facing theme derived from that road label.

#### Scenario: User reads a Dayun with two phase results
- **WHEN** a Dayun includes both `qian_road` and `hou_road`
- **THEN** the interface SHALL show the front-five phase road, calendar range, and theme
- **AND** it SHALL show the back-five phase road, calendar range, and theme
- **AND** it SHALL keep the governing Gan/Zhi and Jin Bu Huan rating as secondary supporting information.

### Requirement: Phase guidance remains distinct from the ten-year composite conclusion
The system SHALL preserve `road_label` as the only ten-year composite road conclusion while describing phase results as phase road conditions.

#### Scenario: Composite and phase roads differ
- **WHEN** a Dayun has composite `泥路` and both phase road labels are `施工路段`
- **THEN** the interface SHALL identify `泥路` as `十年综合路况`
- **AND** it SHALL identify each `施工路段` as the respective five-year phase road
- **AND** it SHALL NOT represent either phase road as the ten-year composite result.
