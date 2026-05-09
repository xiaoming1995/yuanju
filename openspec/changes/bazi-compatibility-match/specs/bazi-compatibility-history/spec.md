## ADDED Requirements

### Requirement: Logged-in users can list and reopen their compatibility readings
The system SHALL provide a private history for compatibility readings that belongs to the authenticated user who created them.

#### Scenario: List compatibility history
- **WHEN** an authenticated user requests compatibility history
- **THEN** the system returns only compatibility readings created by that user
- **AND** each history item includes enough summary data to distinguish one reading from another

#### Scenario: Open compatibility history detail
- **WHEN** an authenticated user requests a specific compatibility reading detail that they own
- **THEN** the system returns both participant snapshots, the structured analysis result, and the latest saved compatibility report if one exists

### Requirement: Compatibility participant data is isolated from normal single-chart history
The system SHALL keep compatibility participant snapshots separate from the existing single-chart history so a partner profile does not appear as the user's own natal chart record.

#### Scenario: Partner profile is not shown in normal history
- **WHEN** a user creates a compatibility reading
- **THEN** the partner participant data is saved under compatibility resources
- **AND** it does not create a new entry in `GET /api/bazi/history`

#### Scenario: Compatibility history remains independently addressable
- **WHEN** the frontend renders the compatibility area
- **THEN** it can navigate to compatibility history and detail routes without reusing single-chart history identifiers

### Requirement: Compatibility readings enforce ownership access control
The system SHALL prevent users from accessing compatibility readings created by other accounts.

#### Scenario: Reject cross-user compatibility detail access
- **WHEN** a user requests a compatibility reading ID that belongs to another account
- **THEN** the system denies access
- **AND** no participant snapshot or report content is returned
