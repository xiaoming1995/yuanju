## ADDED Requirements

### Requirement: Past-events page SHALL expose generation state clearly
The past-events page SHALL distinguish deterministic year readiness from AI dayun summary generation states in user-facing UI.

#### Scenario: Page shows deterministic years and AI progress separately
- **WHEN** the past-events page has loaded deterministic year data
- **AND** one or more dayun AI summaries are still generating
- **THEN** the page SHALL indicate that year signals are ready
- **AND** the page SHALL separately indicate that AI dayun commentary is still generating

#### Scenario: Page shows interrupted dayun summaries as retryable
- **WHEN** a dayun AI summary stream fails, stalls, or is interrupted
- **THEN** the affected dayun segment SHALL render an interrupted state
- **AND** the affected dayun segment SHALL provide a retry action for that segment
- **AND** unrelated completed dayun segments SHALL remain readable

#### Scenario: Page shows generated or cached summaries as complete
- **WHEN** a dayun segment has AI summary content from cache or stream completion
- **THEN** the segment SHALL render a complete/generated state
- **AND** the segment SHALL NOT continue to show a loading or retry-only state

### Requirement: Past-events page SHALL make the current period easy to find
The past-events page SHALL make the current dayun and current year discoverable without requiring the user to scan the entire timeline from the beginning.

#### Scenario: Current dayun is present
- **WHEN** the loaded dayun metadata contains the dayun segment that includes the current year
- **THEN** the page SHALL visually identify the current dayun segment
- **AND** the page SHALL provide a top-level cue, highlight, or navigation affordance that points the user to that segment

#### Scenario: Current year is present
- **WHEN** a year card matches the current year
- **THEN** the year card SHALL be visually distinguishable from surrounding years
- **AND** the distinction SHALL preserve readability on desktop and mobile widths

### Requirement: Future dayun controls SHALL explain the two-step generation pattern
Future dayun segments SHALL keep the existing two-step reveal pattern while making each step's effect clear to the user.

#### Scenario: User reveals future year signals
- **WHEN** a future dayun segment is folded
- **AND** the user clicks the reveal action
- **THEN** the segment SHALL show deterministic year signals without starting an AI generation request
- **AND** the UI SHALL make clear that AI commentary has not yet been generated for that segment

#### Scenario: User generates AI commentary for expanded future segment
- **WHEN** a future dayun segment is expanded and has no cached AI summary
- **AND** the user clicks the AI generation action
- **THEN** the frontend SHALL call the dayun-summary-stream endpoint with only that segment's `dayun_index`
- **AND** the segment SHALL show a generating state until it receives content or an error

### Requirement: Year cards SHALL prioritize readable conclusions while preserving technical detail
Past-events year cards SHALL lead with readable narrative and meaningful signal chips, while keeping technical evidence available as secondary detail.

#### Scenario: Year card has narrative and evidence
- **WHEN** a year card has narrative text and technical evidence summary
- **THEN** the default card presentation SHALL show the readable narrative before dense evidence
- **AND** the technical evidence SHALL remain accessible without replacing the narrative

#### Scenario: Year card has multiple professional details
- **WHEN** a year card includes ten-god power, dayun phase, signal evidence, or similar professional details
- **THEN** those details SHALL not overlap, clip, or dominate the primary narrative on desktop or mobile widths

### Requirement: Past-events signal chips SHALL render meaningful known and unknown signals
The past-events page SHALL render visible chips for meaningful signal types produced by the backend, including newly added algorithm signals.

#### Scenario: Year contains a `夹拱` signal
- **WHEN** a past-events year includes a signal whose type represents `夹拱`
- **THEN** the year card SHALL render a visible `夹拱` chip
- **AND** the chip SHALL be displayed consistently with other event signal chips

#### Scenario: Year contains an unmapped meaningful signal
- **WHEN** a past-events year includes a meaningful signal type that is not in the frontend label map
- **THEN** the year card SHALL render a conservative readable fallback chip instead of silently dropping the signal
