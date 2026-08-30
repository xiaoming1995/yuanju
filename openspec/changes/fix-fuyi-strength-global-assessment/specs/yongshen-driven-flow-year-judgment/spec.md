## MODIFIED Requirements

### Requirement: BaziResult exposes explicit favorable and adverse 十神 lists
The system SHALL compute and surface, on every `BaziResult`, a list of favorable 十神 names, a list of adverse 十神 names, and a confidence band derived from the global Fuyi strength assessment. 调候 urgency SHALL remain available to interpretation but SHALL NOT replace the global Fuyi strength or favorable/adverse elements used to derive these lists.

#### Scenario: 身旺 命主 produces 泄/克/耗 favorable list
- **WHEN** a `BaziResult` has global Fuyi `strength_level` of `strong` or `vstrong`
- **AND** global Fuyi favorable and adverse five-element strings are non-empty
- **THEN** `FavorableShishen` SHALL contain {食神, 伤官, 偏财, 正财, 正官, 七杀}
- **AND** `AdverseShishen` SHALL contain {比肩, 劫财, 偏印, 正印}

#### Scenario: 身弱 命主 produces 生/扶 favorable list
- **WHEN** a `BaziResult` has global Fuyi `strength_level` of `weak` or `vweak`
- **AND** global Fuyi favorable and adverse five-element strings are non-empty
- **THEN** `FavorableShishen` SHALL contain {偏印, 正印, 比肩, 劫财}
- **AND** `AdverseShishen` SHALL contain {食神, 伤官, 偏财, 正财, 正官, 七杀}

#### Scenario: 中和 命主 produces soft-confidence empty list
- **WHEN** a `BaziResult` has global Fuyi `strength_level` of `neutral`
- **THEN** `FavorableShishen` SHALL be empty
- **AND** `AdverseShishen` SHALL be empty
- **AND** `ShishenConfidence` SHALL be `soft`

#### Scenario: Confidence band reflects strength tier
- **WHEN** computing `ShishenConfidence`
- **THEN** the band SHALL be `hard` for `vstrong`/`vweak`, `medium` for `strong`/`weak`, and `soft` for `neutral`
