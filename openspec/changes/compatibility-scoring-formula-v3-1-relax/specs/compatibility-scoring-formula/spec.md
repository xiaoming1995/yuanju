# compatibility-scoring-formula Specification

## MODIFIED Requirements

### Requirement: Zodiac module (year-zhi liuhe / sanhe + same-element + sheng)
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

### Requirement: Day pillar module (four tiers)
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

### Requirement: Eight-chars module (per-pillar 0/3/5/10)
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

### Requirement: Analysis version tag (v3.1)
The analysis_version field of new CompatibilityReading records written by this engine SHALL be "v3.1"; v1/v2/v3 records remain unchanged and renderable through their respective frontend paths.

#### Scenario: New reading written
- **WHEN** CreateCompatibilityReading runs after this change
- **THEN** the stored row SHALL have analysis_version = "v3.1"

#### Scenario: Legacy v3 record read
- **WHEN** a v3 record is fetched
- **THEN** the API SHALL surface its v3 dimension_scores (zodiac/nayin/day_pillar/eight_chars) as-is
- **AND** overall_score SHALL remain whatever was originally stored
