## ADDED Requirements

### Requirement: Composite road and phase prompts have distinct user-facing meanings
The system SHALL present a Dayun `road_label` as the `十年综合路况` and SHALL present `qian_road` and `hou_road` as `金不换阶段提示`, not as equivalent road-grade results.

#### Scenario: Composite result and phase ratings differ
- **WHEN** a Dayun has a composite `road_label` of `泥路` and both Jin Bu Huan phase ratings are `凶`
- **THEN** the interface SHALL identify `泥路` as the ten-year composite road condition
- **AND** it SHALL identify each phase as a Jin Bu Huan prompt with its respective time scope
- **AND** it SHALL NOT describe the two phase results as the overall road condition.

### Requirement: Active Dayun explains the phases in plain language
The system SHALL show the selected or current Dayun's front-five and back-five prompts with their calendar ranges, governing source, and rating when phase data is available.

#### Scenario: User reads a current Dayun with both phase ratings
- **WHEN** the current Dayun includes `qian_road` and `hou_road`
- **THEN** the interface SHALL show a front-five item marked `天干主事` and a back-five item marked `地支主事`
- **AND** each item SHALL identify its five-year range and Jin Bu Huan rating as `吉`, `平`, or `凶`
- **AND** the default view SHALL show a concise explanation without raw score arithmetic.

### Requirement: Road explanation discloses the scoring scope
The system SHALL provide a compact explanation entry point for the active Dayun that states the ten-year road condition is composite and the front/back prompts originate from Jin Bu Huan.

#### Scenario: User asks why labels differ
- **WHEN** the user opens the active Dayun road explanation
- **THEN** the explanation SHALL state that the ten-year road combines multiple deterministic signals
- **AND** it SHALL state that the front and back prompts respectively refer to the Dayun stem and branch
- **AND** it SHALL offer the existing professional evidence route when detailed basis is available.
