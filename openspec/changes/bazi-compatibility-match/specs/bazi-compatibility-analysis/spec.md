## ADDED Requirements

### Requirement: Authenticated users can create a compatibility reading from two complete birth profiles
The system SHALL provide a compatibility analysis entrypoint for authenticated users that accepts two complete birth profiles for a romance/marriage reading. Each profile MUST include year, month, day, hour, gender, calendar type, and leap-month flag when applicable.

#### Scenario: Create compatibility reading successfully
- **WHEN** an authenticated user submits one `self` profile and one `partner` profile with complete birth fields
- **THEN** the system creates a new compatibility reading resource
- **AND** the system calculates both parties' chart snapshots before producing the pair analysis result

#### Scenario: Reject incomplete profile input
- **WHEN** the user omits the birth hour for either participant
- **THEN** the system rejects the request
- **AND** the response explains that the first version requires complete birth time for both participants

### Requirement: Compatibility analysis returns multidimensional results instead of a binary verdict
The system SHALL return a structured compatibility result with one overall level and four core dimensions: `attraction`, `stability`, `communication`, and `practicality`. Each dimension SHALL have a score suitable for UI visualization, and the overall result SHALL be expressed as a level rather than a simple yes/no judgment.

#### Scenario: Successful analysis returns four dimensions
- **WHEN** a compatibility reading is generated successfully
- **THEN** the response contains scores for `attraction`, `stability`, `communication`, and `practicality`
- **AND** the response contains `overall_level` with value `high`, `medium`, or `low`

#### Scenario: Overall level is not exposed as binary compatible/incompatible
- **WHEN** the frontend renders a compatibility reading
- **THEN** it can derive a multidimensional summary from the API response
- **AND** it does not depend on a single boolean field such as `is_match`

### Requirement: Compatibility analysis provides structured evidences with explicit polarity and source
The system SHALL attach structured evidences to each compatibility reading so users can understand why the result was produced. Each evidence item SHALL include at least a dimension, type, polarity, source, title, detail, and weight.

#### Scenario: Positive and negative evidences can coexist
- **WHEN** two charts contain both attraction signals and conflict signals
- **THEN** the reading stores multiple evidence items
- **AND** the evidence items preserve their own polarity instead of being flattened into one conclusion

#### Scenario: Evidence sources are explicit
- **WHEN** the analysis detects signals from areas such as day-master interaction, five-element complement, spouse-palace interaction, spouse-star interaction, GanZhi combinations/clashes, or helper shensha
- **THEN** each evidence item records which source category produced it
- **AND** the detail field explains the specific interaction in human-readable terms
