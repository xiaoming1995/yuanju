## ADDED Requirements

### Requirement: Advanced natal assessment uses global Fuyi evidence
The backend SHALL populate `natal_assessment.fuyi` from the global Fuyi strength assessment and SHALL include the strength level, favorable/adverse elements, and evidence used by the vehicle and road profile.

#### Scenario: New calculation returns a consistent Fuyi assessment
- **WHEN** the Calculate API returns a newly calculated chart
- **THEN** `natal_assessment.fuyi` SHALL contain the same global strength and favorable elements used by its vehicle and road profile

#### Scenario: Historical snapshot has an earlier assessment version
- **WHEN** a saved chart has no natal assessment or has an assessment version older than the current version
- **THEN** the service SHALL recalculate the natal assessment, vehicle profile, and road profile before returning the chart
