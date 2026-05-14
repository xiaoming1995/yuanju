## ADDED Requirements

### Requirement: Dayun ten-god power profile
The system SHALL calculate a structured ten-god power profile for each dayun shown in the past-events module.

#### Scenario: Dayun metadata includes ten-god force
- **WHEN** the past-events years endpoint returns `dayun_meta`
- **THEN** each dayun item SHALL include a ten-god power profile with dominant ten-god, force group, strength label, polarity, plain title, plain text, score, and reason

#### Scenario: Dayun force respects front/back phase
- **WHEN** a dayun has first-five heavenly-stem phase and last-five earthly-branch phase context
- **THEN** the dayun ten-god power calculation SHALL account for both the dayun heavenly stem and earthly branch, with phase-aware weighting available to yearly profiles

### Requirement: Liunian ten-god power profile
The system SHALL calculate a structured ten-god power profile for each liunian year shown in the past-events module.

#### Scenario: Year response includes yearly force
- **WHEN** the past-events years endpoint returns a yearly event item
- **THEN** the item SHALL include a ten-god power profile with dominant ten-god, force group, strength label, polarity, plain title, plain text, score, and reason

#### Scenario: Year force combines liunian and dayun
- **WHEN** a liunian shares the same force group or exact ten-god with its containing dayun
- **THEN** the liunian ten-god power profile SHALL reflect that reinforcement in score and reason

#### Scenario: Year force accounts for natal context
- **WHEN** the natal chart has known day-master strength and yongshen/jishen data
- **THEN** the liunian ten-god power profile SHALL use that context to classify the force as support, pressure, or mixed

### Requirement: Ten-god force grouping
The system SHALL group exact ten-god labels into practical force groups before presenting them to users.

#### Scenario: Raw ten gods map to user-facing groups
- **WHEN** a profile uses 正财 or 偏财
- **THEN** the group SHALL be wealth and the default user-facing label SHALL be 财星

#### Scenario: Official and killing are grouped together
- **WHEN** a profile uses 正官 or 七杀
- **THEN** the group SHALL be official and the default user-facing label SHALL be 官杀

#### Scenario: Seal output and peer groups are supported
- **WHEN** a profile uses 正印, 偏印, 食神, 伤官, 比肩, or 劫财
- **THEN** the profile SHALL map them respectively into 印星, 食伤, or 比劫 force groups

### Requirement: Plain-language interpretation
The system SHALL provide a concise plain-language interpretation for every ten-god power profile.

#### Scenario: Profile avoids terminology-only output
- **WHEN** a profile is serialized for frontend use
- **THEN** it SHALL include `plain_title` and `plain_text` fields that explain the force in everyday language

#### Scenario: Technical reason remains available
- **WHEN** professional evidence is expanded or tests inspect the profile
- **THEN** the profile SHALL expose a reason explaining the main scoring factors without requiring an LLM call

### Requirement: Narrative uses ten-god force as context
The yearly narrative renderer SHALL use ten-god power as contextual tone while preserving stronger event evidence priority.

#### Scenario: Hard event evidence remains dominant
- **WHEN** a year contains hard evidence such as clash, punishment, void, use-god/jishen position hit, or dayun-liunian double hit
- **THEN** the yearly narrative SHALL keep that hard evidence as the dominant theme rather than replacing it with generic ten-god force wording

#### Scenario: Ten-god force explains otherwise generic years
- **WHEN** a year has no stronger concrete event signal but has a meaningful ten-god power profile
- **THEN** the yearly narrative SHALL use the profile to produce a more specific plain-language theme

### Requirement: Dayun summary prompt receives ten-god profiles
The dayun summary generation flow SHALL include ten-god power profiles in the structured input sent to the dayun summary prompt.

#### Scenario: New summary generation sees force context
- **WHEN** a dayun summary is generated or regenerated
- **THEN** the prompt data SHALL include the dayun force profile and each contained liunian force profile

#### Scenario: Cached summaries are not automatically invalidated
- **WHEN** an existing cached dayun summary is present
- **THEN** the system SHALL continue to serve the cached summary unless a separate explicit regeneration path is used

### Requirement: Frontend compact presentation
The past-events frontend SHALL display ten-god power compactly without creating a dense technical panel.

#### Scenario: Dayun header shows dominant force
- **WHEN** a dayun has a ten-god power profile
- **THEN** the dayun header SHALL show a compact dominant-force label such as 主导：财星偏旺

#### Scenario: Year card shows plain force explanation
- **WHEN** a yearly event card has a ten-god power profile
- **THEN** the card SHALL show a short plain-language force explanation near the year title or narrative

#### Scenario: Technical details stay secondary
- **WHEN** users view the default past-events timeline
- **THEN** exact score and detailed reason SHALL NOT dominate the default card layout and SHOULD be reserved for evidence/debug contexts
