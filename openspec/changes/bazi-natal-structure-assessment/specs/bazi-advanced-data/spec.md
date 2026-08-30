## ADDED Requirements

### Requirement: Persisted natal structure assessment data
The backend Bazi result SHALL provide a versioned `natal_assessment` alongside existing advanced data and SHALL preserve legacy fields unchanged.

#### Scenario: New calculation returns advanced data
- **WHEN** the calculation engine returns a Bazi result
- **THEN** it includes a current version of the natal structure assessment without removing or renaming existing use-god, Ming Ge, ten-god, vehicle, or Dayun fields.

#### Scenario: Old saved chart is loaded
- **WHEN** a saved chart snapshot lacks a current natal structure assessment
- **THEN** the report service derives the assessment from the stored chart data, regenerates dependent vehicle and Dayun data, and persists the upgraded snapshot.
