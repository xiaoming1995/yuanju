# compatibility-scoring-formula Specification

## Purpose
Define the deterministic compatibility scoring formula used for bazi relationship readings, including score modules, evidence emission, level thresholds, and analysis version compatibility.
## Requirements
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
The zodiac module SHALL award 50 points when the two year-zhi form 六合 or 三合 (half-sanhe acceptable); 30 points when the two year-zhi have the same 五行 (双生); 20 points when the two year-zhi 五行 stand in a 生 relation (either direction); otherwise 0.

#### Scenario: Year-zhi liuhe hit
- **WHEN** the year-zhi pair is one of {子丑, 寅亥, 卯戌, 辰酉, 巳申, 午未}
- **THEN** zodiac SHALL be 50

#### Scenario: Year-zhi half-sanhe hit
- **WHEN** the year-zhi pair is two distinct members of one sanhe group {申子辰, 亥卯未, 巳酉丑, 寅午戌}
- **THEN** zodiac SHALL be 50

#### Scenario: Year-zhi same-element (double-life)
- **WHEN** the two year-zhi share the same 五行 (亥子 水、寅卯 木、巳午 火、申酉 金、辰戌丑未 土两两) AND the pair is not 六合/三合
- **THEN** zodiac SHALL be 30
- **AND** this applies even when the pair simultaneously forms a 六冲 (e.g. 辰戌, 丑未) — 纯加分制 with no negative deductions

#### Scenario: Year-zhi sheng
- **WHEN** the two year-zhi 五行 stand in a 生 relation (either direction, e.g. 子→寅 水生木) AND the pair is neither 六合/三合 nor same-element
- **THEN** zodiac SHALL be 20

#### Scenario: No hit (相克 / 相冲 with different 五行 / 相穿 / 相害 / 自刑 / 同支)
- **WHEN** none of the above hits
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
The day_pillar module SHALL award:
- 10 (上档 / `day_pillar_upper`) when day-zhi is 六合/三合 AND day-gan is 五合 or 五行相生
- 5 (下档 / `day_pillar_lower`) when day-zhi is 六合/三合 (day-gan 同/克/无关)
- 3 (安慰分 / `day_pillar_safe`) when day-zhi 五行同 (双生) OR 五行相生 (day-gan ignored)
- 0 when day-zhi is 相克 / 相冲 / 相穿 / 相害 / 同支

注：tier 命名 "lower" 保留 v3 原有的 evidence_key 字符串以避免破坏 v3 记录的 evidence 兼容性；v3.1 新增的 3 分档使用独立 key `day_pillar_safe`。

#### Scenario: Upper tier — gan 五合 + zhi 六合/三合
- **WHEN** day-gan pair is in 天干五合 set {甲己, 乙庚, 丙辛, 丁壬, 戊癸} AND day-zhi pair is 六合/三合
- **THEN** day_pillar SHALL be 10

#### Scenario: Upper tier — gan 五行相生 + zhi 六合/三合
- **WHEN** the two day-gan 五行 stand in a 相生 relation (excluding identity) AND day-zhi pair is 六合/三合
- **THEN** day_pillar SHALL be 10

#### Scenario: Lower tier — zhi 六合/三合 alone
- **WHEN** day-zhi pair is 六合/三合 AND no upper-tier 干 condition met
- **THEN** day_pillar SHALL be 5

#### Scenario: Safe tier — zhi same-element or sheng
- **WHEN** day-zhi pair is 五行同 (双生) OR 五行相生 AND not 六合/三合
- **THEN** day_pillar SHALL be 3, regardless of day-gan relation

#### Scenario: zhi 不合（相克 / 相冲 with different 五行 / 同支）
- **WHEN** none of the above scenarios apply
- **THEN** day_pillar SHALL be 0

### Requirement: Eight-chars module (year/month/hour aggregation)
The eight_chars module SHALL score each of the three non-day pillar pairs (year/year, month/month, hour/hour) by the day_pillar rule, sum the three results (0–30) and normalize to [0, 20] via (sum × 2 + 1) / 3 with integer division.

#### Scenario: All three pillars upper tier
- **WHEN** year/month/hour pillar pairs each score 10
- **THEN** eight_chars SHALL be 20

#### Scenario: One pillar safe tier, two not compatible
- **WHEN** sum is 3
- **THEN** eight_chars SHALL be 2

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
The analysis_version field of new CompatibilityReading records written by this engine SHALL be "v3.1"; v1/v2/v3 records remain unchanged and renderable through their respective frontend paths.

#### Scenario: New reading written
- **WHEN** CreateCompatibilityReading runs after this change
- **THEN** the stored row SHALL have analysis_version = "v3.1"

#### Scenario: Legacy v3 record read
- **WHEN** a v3 record is fetched
- **THEN** the API SHALL surface its v3 dimension_scores (zodiac/nayin/day_pillar/eight_chars) as-is
- **AND** overall_score SHALL remain whatever was originally stored
