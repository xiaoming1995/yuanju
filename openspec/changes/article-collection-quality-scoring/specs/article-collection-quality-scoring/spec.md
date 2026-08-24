## ADDED Requirements

### Requirement: Candidate Quality Scoring
The system SHALL calculate an explainable quality score for article collection candidates before admin review.

#### Scenario: Candidate is scored
- **WHEN** a collection provider returns an article candidate
- **THEN** the system SHALL calculate a numeric quality score
- **AND** the system SHALL produce structured human-readable reasons for the score

#### Scenario: Candidate has no external metrics
- **WHEN** a candidate has no WeChat like, favorite, read, or other external interaction metrics
- **THEN** the system SHALL still calculate a quality score from available metadata and content signals
- **AND** the missing external metrics SHALL NOT block collection

### Requirement: Quality Filtering Configuration
The system SHALL allow administrators to configure collection quality filtering.

#### Scenario: Admin updates quality filter settings
- **WHEN** an authenticated administrator updates quality filtering settings
- **THEN** the system SHALL persist whether filtering is enabled
- **AND** the system SHALL persist the minimum quality score and configured source/keyword rules

#### Scenario: Filtering is disabled
- **WHEN** quality filtering is disabled
- **THEN** the system SHALL store candidate articles according to the normal collection rules
- **AND** the system SHALL still persist quality score and reasons for inspection

#### Scenario: Filtering is enabled
- **WHEN** quality filtering is enabled
- **AND** a candidate score is below the configured minimum
- **THEN** the system SHALL skip article insertion
- **AND** the system SHALL record a task item with skipped status, score, and skip reason

### Requirement: Source Rule Controls
The system SHALL support source-level quality rules for collected articles.

#### Scenario: Source is blacklisted
- **WHEN** a candidate source matches a configured source blacklist rule
- **THEN** the system SHALL skip or heavily penalize the candidate according to configuration
- **AND** the quality reason SHALL identify the source rule

#### Scenario: Source is preferred
- **WHEN** a candidate source matches a configured preferred-source rule
- **THEN** the system MAY add source quality points
- **AND** the quality reason SHALL identify the source boost

### Requirement: Quality Visibility In Admin
The system SHALL expose quality score and reasons to administrators.

#### Scenario: Admin views task log items
- **WHEN** an authenticated administrator views collection task item details
- **THEN** the system SHALL show quality score and quality reasons for each available item
- **AND** skipped items SHALL show the skip reason

#### Scenario: Admin reviews article candidates
- **WHEN** an authenticated administrator views article candidates
- **THEN** the system SHALL expose each article quality score
- **AND** the administrator SHALL be able to inspect the scoring reasons on article detail

### Requirement: Yuanju Behavior Signals
The system SHALL support Yuanju-owned behavior signals for article ranking after publication.

#### Scenario: Published article has station behavior
- **WHEN** users view an article, click its original link, or favorite it inside Yuanju
- **THEN** the system MAY use those station-owned behavior signals for hot sorting or future quality ranking
- **AND** these station-owned signals SHALL be separate from unavailable WeChat-side metrics
