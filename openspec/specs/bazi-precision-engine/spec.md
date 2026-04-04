## ADDED Requirements

### Requirement: Precise Astronomical Charting
The system SHALL use the `lunar-go` library for Bazi calculation to ensure Jieqi dates are calculated using true astrological algorithms rather than rough mathematical averages.

#### Scenario: Edge Case Solar Term Switch
- **WHEN** a user is born exactly near a Jieqi transition time
- **THEN** the system computes the exact solar longitude and correctly aligns the year or month pillar accordingly

### Requirement: Hidden Stems Generation
The system SHALL extract and expose the hidden stems (地支藏干) associated with each branch in the four pillars.

#### Scenario: Professional View Display
- **WHEN** the backend returns a Bazi chart
- **THEN** it includes hidden stems data, which the frontend displays in the professional view mode

### Requirement: True Solar Time Correction
The system SHALL accept a longitude parameter during the `calculate` API call to augment the inputted local hour into True Solar Time.

#### Scenario: Longitude Given
- **WHEN** the user provides `longitude` data via the API
- **THEN** the exact True Solar Time is used for calculating the Hour Pillar
- **WHEN** the user does not provide `longitude`
- **THEN** Beijing time is assumed.
