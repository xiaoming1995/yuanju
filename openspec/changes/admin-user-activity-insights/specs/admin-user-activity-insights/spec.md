## ADDED Requirements

### Requirement: Accurate authentication and activity timestamps
The system SHALL record successful user authentication with its channel and timestamp, and SHALL keep last login separate from last tracked activity. Registration SHALL be distinguishable from login, and historical values SHALL remain unknown when no event was recorded.

#### Scenario: Web email login succeeds
- **WHEN** an existing user successfully completes the web email login endpoint
- **THEN** the system records one web login event and updates that user's last login and last activity timestamps

#### Scenario: Miniapp first login creates an account
- **WHEN** a WeChat code successfully creates a new miniapp user and returns a token
- **THEN** the system records registration and one miniapp login event for that user

#### Scenario: Account linking verifies web credentials
- **WHEN** a miniapp user links an existing web account using its email password
- **THEN** the system records account linking for the surviving account and SHALL NOT count the internal credential check as a web login

#### Scenario: Authentication fails or an old user has no events
- **WHEN** authentication fails or an existing user has no recorded login
- **THEN** failure does not change last login and the administrator sees an unknown or never logged in state rather than an inferred time

### Requirement: Bounded core business event tracking
The system SHALL record only successful, named core user actions with a timestamp and minimal reference fields. It SHALL NOT store submitted birth details, password material, report body, full request payload, IP address, or user agent in these events.

#### Scenario: Core action succeeds
- **WHEN** an authenticated user successfully creates a chart or compatibility reading, generates and persists a main bazi or compatibility AI report, or updates profile nickname or avatar
- **THEN** the system records the corresponding action for that user and advances last activity

#### Scenario: Action fails or cached report is read
- **WHEN** a tracked action fails, is still streaming, or an existing cached report is merely read
- **THEN** the system does not record a successful creation or generation event

#### Scenario: Guest chart is created
- **WHEN** an unauthenticated visitor creates a chart
- **THEN** the system retains existing chart behavior without attributing a user activity event

### Requirement: Admin user list and timeline
The system SHALL expose last login and last activity in the administrator user list and SHALL provide an Admin JWT protected, paginated user activity timeline and account summary. List filters SHALL include registration source, account status, never logged in, and recent activity, with time sorting while preserving existing search and default pagination.

#### Scenario: Administrator filters inactive accounts
- **WHEN** an authenticated administrator filters for users who never logged in or were not active in the last 30 days
- **THEN** the list returns matching users with correct total counts and clearly labeled empty timestamps

#### Scenario: Administrator opens user detail
- **WHEN** an authenticated administrator opens an existing user detail and requests its first activity page
- **THEN** the response contains limited account summary and newest events first, with page boundaries and no private chart or report content

#### Scenario: Ordinary user requests admin timeline
- **WHEN** a non-admin requests the user detail or activity endpoint
- **THEN** the request is rejected by the existing administrator authorization boundary

### Requirement: Auditable administrator user actions
The system SHALL record the acting administrator, target user ID, action and timestamp for successful user creation, disable, enable, password reset and deletion. Failed operations SHALL NOT be represented as successful audit actions.

#### Scenario: Administrator disables user
- **WHEN** an administrator successfully disables a user
- **THEN** one audit record identifies the administrator, target UUID, disable action and completion time

#### Scenario: Administrator deletes user
- **WHEN** an administrator successfully deletes a user
- **THEN** the user's activity events are removed and a minimal deletion audit record remains without email or nickname

### Requirement: Event lifecycle and account merge
The system SHALL retain user activity detail for 90 days and administrator action audit for 180 days, while preserving last login and last activity account snapshots. A miniapp-to-web account merge SHALL retain correct event ownership and latest timestamps on the surviving user.

#### Scenario: Retention cleanup runs
- **WHEN** the daily cleanup processes events older than their respective retention periods
- **THEN** expired detail is removed and the user account's last login and activity timestamps remain intact

#### Scenario: Miniapp user merges into web user
- **WHEN** a miniapp account is linked and merged into an existing web account
- **THEN** prior unexpired events belong to the surviving account and its last login and activity reflect the later values from both accounts
