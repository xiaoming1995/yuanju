## ADDED Requirements

### Requirement: Compatibility readings capture relationship context
The system SHALL allow compatibility creation to include the user's current relationship stage and primary relationship question.

#### Scenario: Create reading with relationship context
- **WHEN** an authenticated user creates a compatibility reading with `relationship_stage` and `primary_question`
- **THEN** the system stores those values with the compatibility reading
- **AND** subsequent detail responses include the stored context

#### Scenario: Create reading without relationship context
- **WHEN** an authenticated user creates a compatibility reading without context fields
- **THEN** the system creates the reading successfully
- **AND** the reading behaves as a general relationship judgment

### Requirement: Relationship context uses controlled values
The system SHALL normalize relationship context to controlled values so reports and UI labels remain deterministic.

#### Scenario: Unknown context value is submitted
- **WHEN** the create request contains an unrecognized relationship stage or primary question
- **THEN** the system falls back to `general`
- **AND** it does not reject an otherwise valid compatibility reading

#### Scenario: Detail response exposes normalized context
- **WHEN** a compatibility detail response is returned
- **THEN** the context fields contain normalized enum values
- **AND** the frontend can map them to stable display labels

### Requirement: Compatibility history can summarize relationship context
The system SHALL make relationship context available to compatibility history entries so users can distinguish why each reading was created.

#### Scenario: History item includes context labels
- **WHEN** a user views compatibility history
- **THEN** each item can expose the relationship stage and primary question when available
- **AND** items without context remain viewable with a general fallback label
