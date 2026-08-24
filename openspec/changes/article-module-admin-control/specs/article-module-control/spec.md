## ADDED Requirements

### Requirement: Article Module Availability Setting
The system SHALL persist a global boolean setting that controls whether the article module is running.

#### Scenario: Default setting after migration
- **WHEN** the module availability setting does not yet exist
- **THEN** the system SHALL treat the article module as enabled by default
- **AND** an additive migration SHALL insert the setting with enabled value

#### Scenario: Setting is updated
- **WHEN** an authenticated administrator updates article module availability
- **THEN** the system SHALL persist the new boolean value
- **AND** subsequent user-facing article access decisions SHALL use that value
- **AND** subsequent collection execution decisions SHALL use that value

### Requirement: Admin Module Settings API
The system SHALL provide authenticated admin APIs for reading and updating article module availability.

#### Scenario: Admin reads module setting
- **WHEN** an authenticated administrator requests the article module setting
- **THEN** the system SHALL return whether the article module is enabled

#### Scenario: Admin disables module
- **WHEN** an authenticated administrator submits `module_enabled=false`
- **THEN** the system SHALL save the module as disabled
- **AND** ordinary users SHALL no longer be able to browse article list or detail content
- **AND** scheduled, manual, and retry collection SHALL no longer execute

#### Scenario: Non-admin updates module setting
- **WHEN** a request without a valid Admin JWT attempts to update the article module setting
- **THEN** the system SHALL reject the request with an authentication error

### Requirement: Public Module Settings API
The system SHALL provide a lightweight API for frontend clients to determine whether the article module entry should be shown.

#### Scenario: Frontend reads article settings
- **WHEN** the frontend requests article module settings
- **THEN** the system SHALL return `module_enabled`
- **AND** the response SHALL NOT require ordinary user authentication

#### Scenario: Settings read fails
- **WHEN** the frontend cannot read article module settings due to a transient error
- **THEN** the frontend SHALL avoid showing a broken article entry
- **AND** direct article pages SHALL still rely on backend access control

### Requirement: Closed Module User Response
The system SHALL return a clear closed-module response for ordinary-user article operations when the module is disabled.

#### Scenario: User lists articles while module is disabled
- **WHEN** an ordinary user requests the article list while the article module is disabled
- **THEN** the system SHALL reject the request with HTTP 403
- **AND** the response SHALL identify that the article module is not open

#### Scenario: User opens article detail while module is disabled
- **WHEN** an ordinary user requests article detail while the article module is disabled
- **THEN** the system SHALL reject the request with HTTP 403
- **AND** the system SHALL NOT increment article view count

#### Scenario: User tracks original click while module is disabled
- **WHEN** an ordinary user requests original-link click tracking while the article module is disabled
- **THEN** the system SHALL reject the request with HTTP 403
- **AND** the system SHALL NOT write an original-click log
