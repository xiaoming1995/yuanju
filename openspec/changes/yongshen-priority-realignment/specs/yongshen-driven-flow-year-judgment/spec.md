## ADDED Requirements

### Requirement: BaziResult exposes explicit favorable and adverse 十神 lists
The system SHALL compute and surface, on every `BaziResult`, a list of favorable 十神 names, a list of adverse 十神 names, and a confidence band derived from 调候用神 + 扶抑用神 + 身强弱.

#### Scenario: 身旺 命主 produces 泄/克/耗 favorable list
- **WHEN** a `BaziResult` has `strength_level` of `strong` or `vstrong`
- **AND** `Yongshen` and `Jishen` five-element strings are non-empty
- **THEN** `FavorableShishen` SHALL contain {食神, 伤官, 偏财, 正财, 正官, 七杀}
- **AND** `AdverseShishen` SHALL contain {比肩, 劫财, 偏印, 正印}

#### Scenario: 身弱 命主 produces 生/扶 favorable list
- **WHEN** a `BaziResult` has `strength_level` of `weak` or `vweak`
- **AND** `Yongshen` and `Jishen` five-element strings are non-empty
- **THEN** `FavorableShishen` SHALL contain {偏印, 正印, 比肩, 劫财}
- **AND** `AdverseShishen` SHALL contain {食神, 伤官, 偏财, 正财, 正官, 七杀}

#### Scenario: 中和 命主 produces soft-confidence empty list
- **WHEN** a `BaziResult` has `strength_level` of `neutral`
- **THEN** `FavorableShishen` SHALL be empty
- **AND** `AdverseShishen` SHALL be empty
- **AND** `ShishenConfidence` SHALL be `soft`

#### Scenario: Confidence band reflects strength tier
- **WHEN** computing `ShishenConfidence`
- **THEN** the band SHALL be `hard` for `vstrong`/`vweak`, `medium` for `strong`/`weak`, and `soft` for `neutral`

### Requirement: Past-events dayun summary prompt SHALL inject 喜忌十神 context
The system SHALL include the favorable/adverse 十神 information, gated by confidence band, in the AI prompt used to generate dayun-segment year narratives.

#### Scenario: Hard-confidence命主 receives explicit lists in prompt
- **WHEN** `GenerateDayunSummariesStream` builds the prompt for a `BaziResult` with `ShishenConfidence == "hard"`
- **THEN** the rendered prompt SHALL contain a line naming each item in `FavorableShishen` as `本命喜十神`
- **AND** a line naming each item in `AdverseShishen` as `本命忌十神`

#### Scenario: Soft-confidence命主 receives 调候 guidance instead of list
- **WHEN** `GenerateDayunSummariesStream` builds the prompt for a `BaziResult` with `ShishenConfidence == "soft"`
- **THEN** the rendered prompt SHALL contain a line stating `本命喜忌不显（中和命主），以调候用神...为主`
- **AND** the prompt SHALL NOT contain `本命喜十神` or `本命忌十神` lines

#### Scenario: Medium-confidence命主 receives a softened single-direction hint
- **WHEN** `GenerateDayunSummariesStream` builds the prompt for a `BaziResult` with `ShishenConfidence == "medium"`
- **THEN** the rendered prompt SHALL contain a line stating `本命偏向喜十神：...（中等强度）`

### Requirement: AI-generated artifacts SHALL carry algorithm_version
The system SHALL stamp every newly written `ai_dayun_summaries` and `ai_reports` row with the algorithm version under which it was generated.

#### Scenario: New dayun summary row is stamped with current algorithm version
- **WHEN** `UpsertDayunSummary` writes a new row
- **THEN** the row's `algorithm_version` SHALL equal the current algorithm version constant (initially `v2-yongshen-shishen`)

#### Scenario: New AI report row is stamped with current algorithm version
- **WHEN** the system inserts a new `ai_reports` row
- **THEN** the row's `algorithm_version` SHALL equal the current algorithm version constant

#### Scenario: Historical rows remain readable with null algorithm version
- **WHEN** the system reads an `ai_dayun_summaries` or `ai_reports` row where `algorithm_version` is `NULL`
- **THEN** the row SHALL be treated as algorithm version `v1` (the pre-realignment baseline)
