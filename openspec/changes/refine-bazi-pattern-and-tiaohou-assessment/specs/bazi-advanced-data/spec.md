## ADDED Requirements

### Requirement: Advanced natal data exposes separated Tiaohou evidence
The backend SHALL serialize the day-stem Tiaohou and thermal Tiaohou assessments separately within the versioned natal assessment, while preserving existing raw `tiaohou` data for compatible consumers.

#### Scenario: New chart calculation returns two Tiaohou dimensions
- **WHEN** the Calculate API returns a newly computed chart
- **THEN** its natal assessment SHALL include separately labelled day-stem and thermal Tiaohou results
- **AND THEN** each result SHALL expose its visible and hidden support evidence

#### Scenario: Legacy snapshot is read
- **WHEN** a saved chart lacks the current two-dimensional Tiaohou assessment version
- **THEN** the service SHALL recalculate and persist the updated natal assessment before returning the chart
