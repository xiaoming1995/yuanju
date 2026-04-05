## ADDED Requirements

### Requirement: Engine Support for Dayun Direction Parsing
The backend calculation engine SHALL parse each Dayun's branch (Zhi) into a directional representation (East, South, West, North).

#### Scenario: Dayun Zhi Belongs to East
- **WHEN** the Dayun branch is 寅, 卯, or 辰
- **THEN** it is parsed as "东方木"

#### Scenario: Dayun Zhi Belongs to South
- **WHEN** the Dayun branch is 巳, 午, or 未
- **THEN** it is parsed as "南方火"

#### Scenario: Dayun Zhi Belongs to West
- **WHEN** the Dayun branch is 申, 酉, or 戌
- **THEN** it is parsed as "西方金"

#### Scenario: Dayun Zhi Belongs to North
- **WHEN** the Dayun branch is 亥, 子, or 丑
- **THEN** it is parsed as "北方水"

### Requirement: Jin Bu Huan Dictionary Structure
The system SHALL export a dictionary `jin_bu_huan_dict` that maps "DayGan_MonthZhi" to a structured evaluation containing the poem verse and evaluations mapped to direction keywords or specific branch characters.

#### Scenario: Matching DayGan and MonthZhi
- **WHEN** the user is calculating the Bazi with a specific Day Gan and Month Zhi
- **THEN** the system fetches the correct rule from `jin_bu_huan_dict` and applies it to each of their calculated 10 Dayun pillars.

### Requirement: Jin Bu Huan Evaluation Execution
For each Dayun pillar, the system SHALL check if the Jin Bu Huan dictionary contains an evaluation rule for that pillar's branch mapping (either direct match or directional match).

#### Scenario: Direction Matching
- **WHEN** the Dayun direction is listed in the `GoodDirections` array of the dictionary rule
- **THEN** the system attaches "吉" (good) evaluation parameters to that Dayun
- **WHEN** the Dayun direction is listed in the `BadDirections` array of the dictionary rule
- **THEN** the system attaches "凶" (bad) evaluation parameters to that Dayun

#### Scenario: Specific Branch Override
- **WHEN** the dictionary rule has a `SpecificZhi` definition for the exact Dayun branch (e.g. "申")
- **THEN** explicitly apply that overriding evaluation regardless of the broader directional matching.

### Requirement: API Response Integration
The Bazi calculation API response SHALL include the `JinBuHuan` object under each `DayunItem` when a matched configuration is found.

#### Scenario: Valid Match
- **WHEN** a Jin Bu Huan rule applies and evaluates an active Dayun
- **THEN** the API returns `{ "jin_bu_huan": { "level": "大吉", "keyword": "发财", "text": "..." } }`

#### Scenario: No Match
- **WHEN** the rule has no explicit evaluation for the given Dayun configuration
- **THEN** the `jin_bu_huan` field returns nil/null

### Requirement: Frontend Dayun Timeline rendering
The `DayunTimeline` component SHALL consume the optional `jin_bu_huan` property for each item and display the level and keyword to the user.

#### Scenario: Displaying Jin Bu Huan Badge
- **WHEN** the user views the Dayun timeline and there is Jin Bu Huan data for a pillar
- **THEN** the system renders a badge displaying the evaluation level and keyword beneath the traditional Shishen text.
