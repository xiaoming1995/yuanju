## ADDED Requirements

### Requirement: Admin Article Management Includes Module Control
The admin article management surface SHALL include a control for top-level article module availability.

#### Scenario: Admin opens article management
- **WHEN** an authenticated administrator opens the article management page
- **THEN** the frontend SHALL show the current article module availability
- **AND** the availability control SHALL be visually separate from scheduled collection configuration

#### Scenario: Admin toggles module availability
- **WHEN** an authenticated administrator changes the module availability control
- **THEN** the frontend SHALL call the admin module settings API
- **AND** the page SHALL show the saved state after the API succeeds

### Requirement: Admin Curation Remains Available While Module Is Closed
The admin article management APIs and frontend SHALL remain usable for review and configuration while the article module is disabled.

#### Scenario: Admin lists articles while module is disabled
- **WHEN** an authenticated administrator lists articles while the article module is disabled
- **THEN** the system SHALL return managed article records according to the admin filters

#### Scenario: Admin publishes article while module is disabled
- **WHEN** an authenticated administrator publishes an article while the article module is disabled
- **THEN** the system SHALL apply the normal article status lifecycle
- **AND** ordinary users SHALL still be blocked from viewing the article until the module is enabled

#### Scenario: Admin manages taxonomy while module is disabled
- **WHEN** an authenticated administrator creates, updates, enables, or disables article categories, tags, or keywords while the module is disabled
- **THEN** the system SHALL apply the normal admin taxonomy and keyword behavior

#### Scenario: Admin edits collection configuration while module is disabled
- **WHEN** an authenticated administrator updates collection configuration while the module is disabled
- **THEN** the system SHALL save the collection configuration
- **AND** the saved collection configuration SHALL NOT cause collection to run until the module is enabled
