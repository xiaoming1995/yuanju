## ADDED Requirements

### Requirement: Summary-first AI report reading
Structured Bazi AI reports MUST present a concise digest before long-form chapters.

#### Scenario: User opens a result with a structured AI report
- **WHEN** the report section renders
- **THEN** the page shows a report digest with overall summary, yongshen, jishen, and an actionable reading cue
- **AND** the digest appears before chapter content

### Requirement: Expandable mobile report chapters
Structured Bazi AI report chapters MUST be readable as expandable sections on mobile.

#### Scenario: User scans report chapters on mobile
- **WHEN** report chapters are available
- **THEN** each chapter is rendered with an expandable heading
- **AND** the visible heading/brief helps the user decide whether to open the full content

### Requirement: Report terminology support
The report area MUST include compact terminology explanations for common Bazi concepts.

#### Scenario: User sees technical terms in the report
- **WHEN** the report section renders
- **THEN** the page exposes concise explanations for terms such as 用神, 忌神, 格局, and 大运

### Requirement: Report action bar
The result page MUST provide follow-up actions after or near the AI report.

#### Scenario: User finishes reading the report
- **WHEN** the user reaches the report action area
- **THEN** they can export/share the report, view history when logged in, continue to past-events analysis when available, or start a new chart

### Requirement: Mobile report safe area
The result page MUST reserve enough bottom space so report actions are not obscured by the mobile bottom navigation.

#### Scenario: User scrolls to the report footer on mobile
- **WHEN** the mobile bottom navigation is visible
- **THEN** report actions remain readable and tappable above the navigation
