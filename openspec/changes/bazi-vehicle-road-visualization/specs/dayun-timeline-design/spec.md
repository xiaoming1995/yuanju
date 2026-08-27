## ADDED Requirements

### Requirement: Timeline displays Dayun road labels
The Dayun timeline SHALL display road-condition information for each Dayun period when road map data is available.

#### Scenario: Desktop timeline has road map data
- **WHEN** the result page renders the desktop Dayun timeline with `dayun_roadmap`
- **THEN** each Dayun card displays the corresponding road label such as `高速路`, `城市主路`, `山路`, `泥路`, or `施工路段`
- **AND** the label remains compact enough to preserve the existing ten-card desktop row behavior

#### Scenario: Mobile timeline has road map data
- **WHEN** the result page renders the mobile Dayun timeline with `dayun_roadmap`
- **THEN** each Dayun card displays its road label without covering the Ganzhi, age range, current badge, or transition markers

### Requirement: Timeline displays front/back road phase
The Dayun timeline SHALL expose the front-five-year and back-five-year road phases for the selected Dayun.

#### Scenario: User selects a Dayun with phase data
- **WHEN** a Dayun period is selected
- **THEN** the Dayun summary strip displays the front-five-year phase and back-five-year phase from the matching road map item
- **AND** the phase explanation remains separate from the existing Liunian card list

### Requirement: Existing timeline interactions remain unchanged
Road-condition display SHALL NOT break existing Dayun timeline interactions.

#### Scenario: User selects a Dayun card after road labels are added
- **WHEN** the user clicks a Dayun card
- **THEN** the selected Dayun becomes active and the Liunian panel updates to that period

#### Scenario: User opens Liuyue detail after road labels are added
- **WHEN** the user clicks a Liunian card
- **THEN** the existing Liuyue drawer opens for that year and Ganzhi

#### Scenario: User opens Shensha annotation after road labels are added
- **WHEN** the user clicks a Shensha chip
- **THEN** the existing Shensha annotation behavior remains available
