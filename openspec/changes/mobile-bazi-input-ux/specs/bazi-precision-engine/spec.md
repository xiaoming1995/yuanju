## MODIFIED Requirements

### Requirement: True Solar Time Correction
The system SHALL accept a longitude parameter during the `calculate` API call to augment the inputted local hour into True Solar Time, and the frontend SHALL expose this correction as an optional advanced calibration choice without making it part of the default basic input path.

#### Scenario: Longitude Given
- **WHEN** the user provides `longitude` data via the API
- **THEN** the exact True Solar Time is used for calculating the Hour Pillar
- **WHEN** the user does not provide `longitude`
- **THEN** Beijing time is assumed.

#### Scenario: User keeps default calibration
- **WHEN** the user submits the Bazi input form without selecting a birth location
- **THEN** the frontend sends the existing default longitude behavior
- **AND** the interface communicates that Beijing time is being used

#### Scenario: User selects birth location calibration
- **WHEN** the user selects a supported birth location from advanced calibration
- **THEN** the frontend sends the mapped longitude value through the existing calculate API field
- **AND** no backend API contract change is required
