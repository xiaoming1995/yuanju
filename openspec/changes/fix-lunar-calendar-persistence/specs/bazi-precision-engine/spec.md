## ADDED Requirements

### Requirement: Calendar Type Persistence
The system SHALL persist the `calendar_type` (`solar` or `lunar`) and `is_leap_month` (boolean) parameters in the `bazi_charts` table when a new chart is created.

#### Scenario: Lunar chart creation with leap month
- **WHEN** a user creates a Bazi chart with `calendar_type=lunar` and `is_leap_month=true`
- **THEN** the system stores `calendar_type='lunar'` and `is_leap_month=true` in the `bazi_charts` record

#### Scenario: Solar chart creation (default)
- **WHEN** a user creates a Bazi chart with `calendar_type=solar` (or empty)
- **THEN** the system stores `calendar_type='solar'` and `is_leap_month=false` in the `bazi_charts` record

### Requirement: History Replay Consistency
The system SHALL use the stored `calendar_type` and `is_leap_month` values when re-calculating a Bazi chart from a historical record, ensuring the result is identical to the original calculation.

#### Scenario: Viewing history of a lunar chart
- **WHEN** a user views the history detail of a chart originally created with `calendar_type=lunar` and `is_leap_month=true`
- **THEN** the system re-calculates using the stored `calendar_type=lunar` and `is_leap_month=true`
- **AND** the four pillars match the original calculation result

#### Scenario: Generating AI report for a lunar chart
- **WHEN** a user requests an AI report for a chart originally created with `calendar_type=lunar`
- **THEN** the system re-calculates using the stored calendar parameters
- **AND** the AI report is based on the correct four pillars

### Requirement: Backward Compatibility with Existing Records
The system SHALL treat existing records (created before this change) as `calendar_type='solar'` and `is_leap_month=false` by default, ensuring no disruption to historical data.

#### Scenario: Legacy chart without calendar fields
- **WHEN** the system reads a `bazi_charts` record that was created before the migration
- **THEN** `calendar_type` defaults to `'solar'` and `is_leap_month` defaults to `false`
- **AND** re-calculation produces the same result as before
