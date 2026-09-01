## ADDED Requirements

### Requirement: Bazi result includes wealth profile
The backend Bazi result SHALL provide an optional deterministic `wealth_profile` alongside existing advanced data while preserving all existing chart, Ten God, natal assessment, vehicle, and Dayun fields.

#### Scenario: New calculation returns advanced wealth data
- **WHEN** the calculation engine returns a Bazi result with deterministic natal inputs
- **THEN** the result includes `wealth_profile` without removing or renaming existing fields such as `natal_assessment`, `vehicle_profile`, `dayun_roadmap`, raw Ten God fields, or Fuyi fields

#### Scenario: Old saved chart is loaded
- **WHEN** a saved chart snapshot lacks `wealth_profile` or contains an outdated deterministic wealth profile
- **THEN** the report service derives the wealth profile from the stored chart data and persists the upgraded snapshot

#### Scenario: Wealth profile cannot be derived
- **WHEN** a partial or corrupt saved result lacks enough deterministic inputs to derive wealth profile data
- **THEN** the service preserves the existing chart result and omits `wealth_profile` rather than returning an invented grade
