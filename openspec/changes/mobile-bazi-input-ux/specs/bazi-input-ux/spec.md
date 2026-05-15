## ADDED Requirements

### Requirement: Mobile-first birth input flow
The system SHALL present Bazi birth information input as a mobile-first flow where basic birth data is visually prioritized over professional calibration options.

#### Scenario: Mobile user opens the Bazi input form
- **WHEN** the user opens the Bazi input form on a mobile viewport
- **THEN** the form shows gender, calendar type, birth date, and birth time as the primary visible inputs
- **AND** professional calibration controls do not interrupt the primary input sequence by default

#### Scenario: Desktop user opens the Bazi input form
- **WHEN** the user opens the Bazi input form on a desktop viewport
- **THEN** the form uses available horizontal space for a compact layout
- **AND** it preserves the same input order and field meanings as the mobile flow

### Requirement: Birth input confirmation summary
The system SHALL display a human-readable summary of the selected birth profile before the user submits the Bazi calculation.

#### Scenario: User changes basic birth fields
- **WHEN** the user changes gender, calendar type, birth date, or birth time
- **THEN** the confirmation summary updates to reflect the selected values

#### Scenario: User chooses lunar leap month
- **WHEN** the user selects a lunar leap month
- **THEN** the confirmation summary explicitly indicates that the selected month is a leap month

#### Scenario: User leaves calibration at default
- **WHEN** the user does not select a birth location
- **THEN** the confirmation summary indicates that the chart will be calculated using Beijing time

### Requirement: Advanced calibration controls
The system SHALL group optional precision controls under an advanced calibration area so ordinary users can complete the form without understanding professional correction details.

#### Scenario: Advanced calibration is collapsed
- **WHEN** the user has not opened advanced calibration
- **THEN** birth location and true solar time controls are not shown as primary form fields

#### Scenario: User opens advanced calibration
- **WHEN** the user opens advanced calibration
- **THEN** the system presents optional controls for birth location based true solar time correction
- **AND** the system preserves the current default behavior when no location is selected

### Requirement: Explicit Zi hour disambiguation
The system SHALL express Zi hour disambiguation as explicit mutually exclusive choices rather than an ambiguous standalone checkbox.

#### Scenario: User selects a non-Zi birth time
- **WHEN** the selected birth time is not Zi hour
- **THEN** the form does not show Zi hour disambiguation controls

#### Scenario: User selects Zi hour
- **WHEN** the selected birth time is Zi hour
- **THEN** the form allows the user to distinguish 23:00-23:59 from 00:00-00:59
- **AND** the labels explain whether the day pillar follows the previous day or the current day

#### Scenario: User submits early Zi hour
- **WHEN** the user chooses the 23:00-23:59 Zi hour option and submits the form
- **THEN** the request maps to the existing early Zi hour API behavior without changing the backend contract

### Requirement: Shared birth profile component behavior
The system SHALL keep basic birth profile input behavior consistent across Bazi chart creation and compatibility matching.

#### Scenario: User enters compatibility participants
- **WHEN** the user fills either participant birth profile in the compatibility form
- **THEN** the basic field order, labels, calendar handling, and mobile layout match the Bazi chart creation form

#### Scenario: Shared component date normalization runs
- **WHEN** the user changes year, month, calendar type, or leap month status
- **THEN** invalid day selections are normalized using the existing calendar rules
