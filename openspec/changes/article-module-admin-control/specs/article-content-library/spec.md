## ADDED Requirements

### Requirement: Article Navigation Respects Module Availability
The frontend SHALL expose the normal "资讯" navigation entry only when the article module is enabled for ordinary users.

#### Scenario: Module is enabled
- **WHEN** the frontend renders normal navigation after reading article module settings
- **AND** the article module is enabled
- **THEN** authenticated users SHALL see the "资讯" navigation entry

#### Scenario: Module is disabled
- **WHEN** the frontend renders normal navigation after reading article module settings
- **AND** the article module is disabled
- **THEN** authenticated users SHALL NOT see the "资讯" navigation entry

### Requirement: Direct Article Routes Handle Closed Module
The frontend SHALL handle direct visits to article pages when the module is disabled.

#### Scenario: User opens article list route while disabled
- **WHEN** an authenticated user directly opens `/articles`
- **AND** the article module is disabled
- **THEN** the frontend SHALL show a module-closed state instead of an article list

#### Scenario: User opens article detail route while disabled
- **WHEN** an authenticated user directly opens `/articles/:id`
- **AND** the backend rejects the request because the article module is disabled
- **THEN** the frontend SHALL show a module-closed state
- **AND** the frontend SHALL NOT show a generic missing-article message

### Requirement: User Article APIs Respect Module Availability
The ordinary-user article APIs SHALL require both ordinary user authentication and enabled article module availability.

#### Scenario: Unauthenticated user requests articles
- **WHEN** a request without a valid ordinary user JWT calls the article list endpoint
- **THEN** the system SHALL reject the request with an authentication error

#### Scenario: Authenticated user requests articles while enabled
- **WHEN** a request with a valid ordinary user JWT calls the article list endpoint
- **AND** the article module is enabled
- **THEN** the system SHALL return published article summaries

#### Scenario: Authenticated user requests articles while disabled
- **WHEN** a request with a valid ordinary user JWT calls the article list endpoint
- **AND** the article module is disabled
- **THEN** the system SHALL reject the request with a module-closed error

