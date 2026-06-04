# compatibility-scoring-formula Specification

## ADDED Requirements

### Requirement: Four-module additive scoring
The compatibility engine SHALL compute a total score 0–100 as the sum of four independent module scores: zodiac (0–50), nayin (0–20), day_pillar (0–10), eight_chars (0–20).

#### Scenario: Module score ranges respected
- **WHEN** AnalyzeCompatibility runs on any pair of BaziResult inputs
- **THEN** each module score SHALL fall within its declared range
- **AND** the total score SHALL equal the arithmetic sum of the four module scores
- **AND** the total score SHALL be between 0 and 100 inclusive

#### Scenario: No negative contributions
- **WHEN** any combination of inputs is evaluated
- **THEN** no module SHALL deduct points
- **AND** unfavorable interactions (six-chong, six-hai, xing, ke nayin) SHALL contribute 0 instead of negative values

### Requirement: Zodiac module (year-zhi liuhe / sanhe only)
The zodiac module SHALL award 50 points when the two year-zhi form 六合 or 三合 (half-sanhe acceptable), otherwise 0.

#### Scenario: Year-zhi liuhe hit
- **WHEN** the year-zhi pair is one of {子丑, 寅亥, 卯戌, 辰酉, 巳申, 午未}
- **THEN** zodiac SHALL be 50

#### Scenario: Year-zhi half-sanhe hit
- **WHEN** the year-zhi pair is two distinct members of one sanhe group {申子辰, 亥卯未, 巳酉丑, 寅午戌}
- **THEN** zodiac SHALL be 50

#### Scenario: No hit (including 双生, six-chong, six-hai, xing)
- **WHEN** none of the above hits, even when 五行相生/同 (双生) applies
- **THEN** zodiac SHALL be 0

### Requirement: Nayin module (year-pillar nayin element)
The nayin module SHALL award 20 points when the two year-pillar nayin 五行 are 相生 or 相同, otherwise 0.

#### Scenario: Nayin 相生 or 相同
- **WHEN** nayin elements have a 生 or 同 relationship
- **THEN** nayin SHALL be 20

#### Scenario: Nayin 相克
- **WHEN** nayin elements have a 克 relationship
- **THEN** nayin SHALL be 0

### Requirement: Day pillar module
The day_pillar module SHALL award 10 (upper tier) when the day-zhi pair is 支合 AND the day-gan pair is 干合 or 干相生; 5 (lower tier) when 支合 alone (regardless of 干 relation among same/克/无关); 0 when 支不合.

#### Scenario: Upper tier — gan 五合 + zhi 合
- **WHEN** day-gan pair is in 天干五合 set {甲己, 乙庚, 丙辛, 丁壬, 戊癸} AND day-zhi pair is 支合
- **THEN** day_pillar SHALL be 10

#### Scenario: Upper tier — gan 五行相生 + zhi 合
- **WHEN** the two day-gan 五行 stand in a 相生 relation (excluding identity) AND day-zhi pair is 支合
- **THEN** day_pillar SHALL be 10

#### Scenario: Lower tier — zhi 合 alone (gan 同/克/无关)
- **WHEN** day-zhi pair is 支合 AND no upper-tier 干 condition met
- **THEN** day_pillar SHALL be 5

#### Scenario: zhi 不合
- **WHEN** day-zhi pair is not 支合
- **THEN** day_pillar SHALL be 0, regardless of any 干 relation

### Requirement: Eight-chars module (year/month/hour aggregation)
The eight_chars module SHALL score each of the three non-day pillar pairs (year/year, month/month, hour/hour) by the day_pillar rule, sum the three results (0–30) and normalize to [0, 20] via (sum × 2 + 1) / 3.

#### Scenario: All three pillars upper tier
- **WHEN** year/month/hour pillar pairs each score 10
- **THEN** eight_chars SHALL be 20

#### Scenario: One pillar upper tier, two not compatible
- **WHEN** sum is 10
- **THEN** eight_chars SHALL be 7

### Requirement: Overall level threshold
The overall_level SHALL map from total score: ≥ 80 → "high"; 60–79 → "medium"; < 60 → "low".

#### Scenario: Boundary at 80
- **WHEN** total score is exactly 80
- **THEN** overall_level SHALL be "high"

#### Scenario: Boundary at 60
- **WHEN** total score is exactly 60
- **THEN** overall_level SHALL be "medium"

#### Scenario: Boundary at 59
- **WHEN** total score is 59
- **THEN** overall_level SHALL be "low"

### Requirement: Evidence list only contains positive hits
The evidences list SHALL contain at most six items: one each for zodiac/nayin/day_pillar hits and up to three for eight_chars (year, month, hour). All evidence polarity SHALL be "positive"; unhit modules contribute no evidence.

#### Scenario: All four modules score positive
- **WHEN** zodiac > 0 AND nayin > 0 AND day_pillar > 0 AND all three eight_chars sub-pillars hit
- **THEN** evidences SHALL contain exactly 6 entries

#### Scenario: No module hits
- **WHEN** all four module scores are 0
- **THEN** evidences SHALL be an empty list

### Requirement: Analysis version tag
The analysis_version field of new CompatibilityReading records written by this engine SHALL be "v3"; v1/v2 records remain unchanged and renderable through the legacy frontend path.

#### Scenario: New reading written
- **WHEN** CreateCompatibilityReading runs after this change
- **THEN** the stored row SHALL have analysis_version = "v3"

#### Scenario: Legacy record read
- **WHEN** a v1 or v2 record is fetched
- **THEN** the API SHALL surface its dimension_scores with the original 4-dim keys (attraction/stability/communication/practicality)
- **AND** overall_score SHALL be 0 for these legacy records
