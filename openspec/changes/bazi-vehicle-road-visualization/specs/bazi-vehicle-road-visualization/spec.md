## ADDED Requirements

### Requirement: Natal chart vehicle profile
The system SHALL compute a deterministic vehicle profile for each newly calculated Bazi result using existing algorithmic signals.

#### Scenario: New Bazi calculation returns vehicle profile
- **WHEN** the backend calculates a Bazi result
- **THEN** the result includes `vehicle_profile` with `grade`, `score`, `vehicle_type`, `summary`, `tags`, and `evidences`
- **AND** `score` is an integer from 0 to 100
- **AND** `grade` is one of `S`, `A`, `B`, `C`, or `D`

#### Scenario: Vehicle profile is evidence-backed
- **WHEN** `vehicle_profile` is present
- **THEN** its `evidences` include source labels and score impacts for the major contributing signals
- **AND** the evidence references deterministic inputs such as Ming Ge, day-master strength, Tiaohou, Yongshen/Jishen, Ten God confidence, or natal risk modifiers

### Requirement: Natal chart tiers prioritize Tiaohou and Fuyi Yongshen
The system SHALL determine the natal vehicle tier from a Tiaohou-first, Fuyi-second evaluation instead of a parallel positive-addition model.

#### Scenario: Urgent Tiaohou condition is unresolved
- **WHEN** an extreme cold or hot natal chart lacks the required Tiaohou relief in its exposed or hidden stems
- **THEN** its vehicle grade MUST NOT exceed `C`
- **AND** the vehicle evidence identifies the unresolved Tiaohou condition

#### Scenario: Urgent Tiaohou condition is only hidden
- **WHEN** an extreme cold or hot natal chart has the required Tiaohou relief only in hidden stems
- **THEN** its vehicle grade MUST NOT exceed `A`

#### Scenario: Tiaohou is resolved and Fuyi Yongshen is effective
- **WHEN** the urgent Tiaohou condition is resolved, or no urgent Tiaohou condition exists, and the independently derived Fuyi Yongshen is exposed or rooted without dominant exposed Jishen
- **THEN** the profile MAY qualify for `A` or `S` according to its remaining deterministic evidence
- **AND** the evidence identifies the Fuyi Yongshen assessment separately from Tiaohou

#### Scenario: Inference confidence does not inflate tier
- **WHEN** a profile is scored
- **THEN** Ten God inference confidence MAY be displayed as diagnostic evidence
- **BUT** it MUST NOT add points to the natal vehicle score

### Requirement: Primary vehicle class follows natal tier
The system SHALL derive the primary vehicle class from the final S/A/B/C/D natal tier, not from Ming Ge.

#### Scenario: Vehicle class is deterministic for every grade
- **WHEN** a profile has a final natal grade
- **THEN** its primary `vehicle_type` MUST be one of the grade-mapped classes: 超跑级座驾, 高性能车级座驾, 标准轿车级座驾, 实用 MPV 级座驾, or 基础代步单车级

#### Scenario: Ming Ge does not change primary vehicle class
- **WHEN** two profiles have the same final natal grade but different Ming Ge values
- **THEN** they MUST have the same primary `vehicle_type`

#### Scenario: Professional view exposes driving style separately
- **WHEN** a professional-mode vehicle profile has Ming Ge data
- **THEN** the profile MAY display a `driving_style` explanation derived from the Ming Ge
- **AND** the explanation MUST state that it is a handling or usage tendency, not a basis for the grade or primary vehicle class

### Requirement: Dayun road map
The system SHALL compute deterministic road-condition data for each Dayun period in a newly calculated Bazi result.

#### Scenario: New Bazi calculation returns Dayun road map
- **WHEN** the backend calculates a Bazi result with Dayun periods
- **THEN** the result includes `dayun_roadmap`
- **AND** each road item includes `dayun_index`, `gan_zhi`, `score`, `road_type`, `qian_road`, `hou_road`, `summary`, `tags`, and `evidences`
- **AND** each road item maps to one existing Dayun period by `dayun_index`

#### Scenario: Road phases split front and back five years
- **WHEN** a Dayun road item is generated
- **THEN** `qian_road` describes the front five-year phase
- **AND** `hou_road` describes the back five-year phase
- **AND** the phase evidence includes Jin Bu Huan front/back ratings when available

### Requirement: Vehicle and road labels remain metaphorical
The system SHALL present vehicle grades and road types as explanatory metaphors rather than deterministic social rank or guaranteed life outcomes.

#### Scenario: User views vehicle and road labels
- **WHEN** the frontend or AI report displays a vehicle grade or road type
- **THEN** the copy frames the result as configuration, ease-of-use, road support, or risk-control guidance
- **AND** the copy does not use absolute claims such as guaranteed wealth, guaranteed failure, fixed social class, or moral worth

### Requirement: Vehicle grades and road labels are understandable to ordinary users
The system SHALL provide an in-context explanation of the vehicle grade scale and Dayun road labels whenever a vehicle profile is displayed.

#### Scenario: User views a vehicle profile
- **WHEN** the result page displays `vehicle_profile`
- **THEN** the primary grade name is a Chinese natal-chart tier label and the `S/A/B/C/D` value is only an algorithmic marker
- **AND** the page visibly displays every vehicle grade and its ordinary-user explanation without requiring interaction
- **AND** the page identifies the current result's grade in that visible scale
- **AND** the page provides a secondary explanation of the relationship between natal vehicle and Dayun road

#### Scenario: User views a professional vehicle profile
- **WHEN** the result page is in professional mode and `driving_style` is present
- **THEN** it displays the driving style inside the professional evidence area
- **AND** ordinary mode does not use the Ming Ge-derived style as the primary vehicle name

#### Scenario: User views the current Dayun road
- **WHEN** the result page resolves a current Dayun road
- **THEN** the current-road card displays a one-line ordinary-user explanation of that road condition
- **AND** the explanation does not represent the road condition as a guaranteed outcome

### Requirement: Current road can be resolved
The system SHALL identify the current Dayun road for the current Gregorian year when Dayun road data is available.

#### Scenario: Current year falls within a Dayun period
- **WHEN** the current Gregorian year is between a Dayun period's `start_year` and `end_year`
- **THEN** the frontend can display the matching `dayun_roadmap` item as the current road condition

#### Scenario: Current year is outside returned Dayun periods
- **WHEN** no Dayun period contains the current Gregorian year
- **THEN** the frontend falls back to the first available Dayun road item or hides the current-road card without breaking the result page
