## MODIFIED Requirements

### Requirement: Article Audit Trail
The system SHALL record administrator, system action, and timestamp audit events for publish, automatic publish, reject, takedown, and delete actions.

#### Scenario: Admin publishes article
- **WHEN** an authenticated administrator publishes an article
- **THEN** the system SHALL write an audit event containing the article ID, admin ID, action, prior status, new status, and timestamp

#### Scenario: System auto-publishes article
- **WHEN** scheduled collection automatically publishes an eligible article
- **THEN** the system SHALL write an audit event containing the article ID, no admin ID, an automatic publish action, prior status, new status, timestamp, and explanatory note

#### Scenario: Admin takes down article
- **WHEN** an authenticated administrator takes down an article
- **THEN** the system SHALL write an audit event containing the article ID, admin ID, action, prior status, new status, and timestamp

## ADDED Requirements

### Requirement: Auto-Published Article Visibility
The admin curation workflow SHALL make automatically published articles distinguishable from manually published articles.

#### Scenario: Admin reviews article audit history
- **WHEN** an authenticated administrator inspects an auto-published article audit event
- **THEN** the system SHALL identify the publication as automatic
- **AND** the audit note SHALL include the quality threshold basis

#### Scenario: Admin filters published articles
- **WHEN** an authenticated administrator lists published articles
- **THEN** auto-published articles SHALL appear in the normal published article list
- **AND** they SHALL remain eligible for normal takedown and delete actions
