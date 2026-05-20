## ADDED Requirements

### Requirement: Past-events page SHALL auto-generate only past + current dayun segments on initial load

The system SHALL determine which dayun segments to auto-generate based on whether the segment's start age is at or before the user's current age. Segments whose start age exceeds the current age SHALL render as folded (collapsed) and SHALL NOT trigger AI generation until the user explicitly requests it.

#### Scenario: 30-year-old user (1995-born, present year 2026) opens past-events page
- **WHEN** the past-events page loads
- **AND** the user's current age is 31
- **AND** dayuns 1-4 have start_age values of 0, 9, 19, 29 respectively
- **AND** dayuns 5-9 have start_age values of 39, 49, 59, 69, 79
- **THEN** dayuns 1, 2, 3, 4 SHALL be auto-generated (expanded, scheduled for AI calls)
- **AND** dayuns 5, 6, 7, 8, 9 SHALL render folded with an `[展开 ▼]` affordance

#### Scenario: Very young user (5 years old) opens past-events page
- **WHEN** the user's current age is 5
- **AND** only dayun 1 (start_age 0) is at or before the current age
- **THEN** only dayun 1 SHALL be auto-generated
- **AND** all later dayuns SHALL render folded

### Requirement: Folded dayun segments SHALL support a two-click reveal pattern

The system SHALL keep future dayun segments collapsed by default. The first click on `[展开 ▼]` SHALL reveal the segment's year chips without any AI call. The second click on `生成本段 AI 批语` SHALL trigger an AI call for that segment alone.

#### Scenario: User expands a folded future dayun
- **WHEN** the user clicks `[展开 ▼]` on a folded dayun segment
- **THEN** the segment SHALL unfold to display the year chips (algorithm-computed signals)
- **AND** no API call SHALL be initiated
- **AND** an `[生成本段 AI 批语]` action button SHALL render at the bottom of the segment

#### Scenario: User triggers AI generation for an expanded future dayun
- **WHEN** the user clicks `生成本段 AI 批语` on an expanded but uncached dayun segment
- **THEN** the frontend SHALL call the dayun-summary-stream endpoint with the explicit `dayun_indexes: [N]` for that segment only
- **AND** the action button SHALL be replaced by a loading state during generation
- **AND** the generated narratives SHALL appear inline upon SSE completion

#### Scenario: Future dayun with existing cache renders expanded
- **WHEN** a future dayun has a row in `ai_dayun_summaries` (cache hit from prior session)
- **THEN** the segment SHALL render expanded with narratives inline
- **AND** the `[生成本段 AI 批语]` button SHALL NOT appear

### Requirement: The dayun-summary-stream endpoint SHALL accept an optional dayun_indexes filter

The system SHALL extend `POST /api/bazi/past-events/dayun-summary-stream/:chart_id` to accept an optional JSON body `{"dayun_indexes": [N, ...]}`. When the body is empty or missing, the server SHALL compute the auto-gen list. When the array is non-empty, the server SHALL stream only those segments.

#### Scenario: Request with empty body uses default list
- **WHEN** the client sends an empty body to the endpoint
- **AND** the chart owner is currently 31 years old
- **THEN** the server SHALL compute the list of dayun indexes whose start_age ≤ 31
- **AND** the SSE stream SHALL emit only those dayuns

#### Scenario: Request with explicit indexes streams only those
- **WHEN** the client sends `{"dayun_indexes": [5]}` 
- **THEN** the server SHALL stream a single SSE item for dayun_index 5
- **AND** the server SHALL NOT process any other dayun

#### Scenario: Request with empty indexes array is treated as default
- **WHEN** the client sends `{"dayun_indexes": []}`
- **THEN** the server SHALL behave identically to an empty body (compute default list)

### Requirement: AI prompt YearsData JSON SHALL be compressed before being sent to the model

The system SHALL apply token-saving transforms to the YearsData JSON used in the AI prompt. The compression SHALL NOT affect the persisted signals, the SSE response, or the frontend payload.

#### Scenario: ten_god_power.plain_title and plain_text are dropped from prompt JSON
- **WHEN** YearsData is serialized for the AI prompt
- **THEN** each year's `ten_god_power` object SHALL NOT contain `plain_title` or `plain_text` keys
- **AND** the persisted `ai_dayun_summaries.years` SHALL be unchanged from its current schema

#### Scenario: Phase metadata fields are dropped from prompt JSON
- **WHEN** YearsData is serialized for the AI prompt
- **THEN** each year entry SHALL NOT contain `year_in_dayun`, `dayun_phase`, or `dayun_phase_level` keys

#### Scenario: Parenthetical asides are stripped from signal.evidence
- **WHEN** a `signal.evidence` string contains a Chinese full-width parenthetical `（…）` of length ≤ 30 characters
- **THEN** the parenthetical SHALL be removed in the prompt-serialized version
- **AND** the same stripping SHALL apply to half-width `(...)` of the same length constraint

#### Scenario: Signal evidence without parentheticals is unchanged
- **WHEN** a `signal.evidence` string contains no parenthetical asides
- **THEN** the prompt-serialized version SHALL equal the original

### Requirement: New AI-generated rows SHALL be stamped with algorithm version v3-progressive-compressed

The system SHALL update `repository.CurrentAlgorithmVersion` to `"v3-progressive-compressed"`. Newly generated dayun summary rows SHALL carry this version, while older rows remain at their original version values for cohort analysis.

#### Scenario: New dayun summary row carries v3 version
- **WHEN** `UpsertDayunSummary` writes a row after this change ships
- **THEN** the row's `algorithm_version` SHALL equal `"v3-progressive-compressed"`

#### Scenario: Pre-v3 rows are honored as-is by the frontend
- **WHEN** the frontend renders a dayun with `algorithm_version = "v2-yongshen-shishen"`
- **THEN** the segment SHALL render expanded with the cached narratives
- **AND** no re-generation prompt SHALL be shown
