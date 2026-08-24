## ADDED Requirements

### Requirement: Collection Pipeline Respects Module Availability
The article collection pipeline SHALL require the top-level article module switch to be enabled before running scheduled, manual, or retry collection.

#### Scenario: Module disabled and collection enabled
- **WHEN** the article module is disabled for ordinary users
- **AND** scheduled collection is enabled and due
- **THEN** the scheduler SHALL NOT run article collection
- **AND** the scheduler SHALL NOT create a collection task
- **AND** the scheduler SHALL NOT update collection `last_run_at`

#### Scenario: Module enabled and collection disabled
- **WHEN** the article module is enabled for ordinary users
- **AND** scheduled collection is disabled
- **THEN** the scheduler SHALL NOT run scheduled article collection

#### Scenario: Admin manually collects while module disabled
- **WHEN** an authenticated administrator triggers manual collection while the article module is disabled
- **THEN** the system SHALL reject manual collection with a module-closed response
- **AND** the system SHALL NOT create a collection task

#### Scenario: Admin retries collection while module disabled
- **WHEN** an authenticated administrator retries a failed collection task while the article module is disabled
- **THEN** the system SHALL reject retry collection with a module-closed response
- **AND** the system SHALL NOT create a retry collection task

### Requirement: Auto-Published Articles Remain Hidden While Module Is Closed
The collection pipeline SHALL preserve normal auto-publish status behavior for articles that were already published before the module was closed, while user visibility is still controlled by module availability.

#### Scenario: Previously auto-published article while module disabled
- **WHEN** an article was automatically published before the article module was disabled
- **AND** the article module is now disabled for ordinary users
- **THEN** the article SHALL have published status
- **AND** ordinary users SHALL NOT be able to see it until the article module is enabled
