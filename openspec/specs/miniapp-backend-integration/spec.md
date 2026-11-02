# miniapp-backend-integration Specification

## Purpose
TBD - created by archiving change miniapp-backend-user-integration. Update Purpose after archive.
## Requirements
### Requirement: Real chart calculation
The miniapp SHALL submit complete birth inputs to the existing backend calculation API and SHALL display the returned result for that input.

#### Scenario: Successful calculation
- **WHEN** an authenticated user submits a valid date, hour, gender, and optional chart name
- **THEN** the result page SHALL show only the returned chart and retain its chart ID

#### Scenario: Failure
- **WHEN** calculation fails or returns no saved chart ID
- **THEN** the miniapp SHALL show an error/retry state and SHALL NOT display the fixed sample as the submitted result

### Requirement: Real account and history
The miniapp SHALL use ordinary user authentication for profile, chart history, detail and delete operations.

#### Scenario: First WeChat login
- **WHEN** a user taps WeChat login for the first time while miniapp registration is enabled
- **THEN** the miniapp SHALL enter a new account without requiring web credentials
- **THEN** the miniapp SHALL offer optional avatar and nickname selection using the platform's user-initiated controls

#### Scenario: Optional web account binding
- **WHEN** a signed-in miniapp user chooses web account binding in Mine
- **THEN** the miniapp SHALL collect web credentials, explain that records will merge, and use the returned web-user JWT after a successful merge

#### Scenario: Authenticated history
- **WHEN** a logged-in user opens history or Mine
- **THEN** the miniapp SHALL show server-owned records and counts, including empty and paginated states

#### Scenario: Delete chart
- **WHEN** the user confirms deletion of an owned chart
- **THEN** the backend SHALL delete it and both miniapp history summaries SHALL refresh

#### Scenario: Token expired
- **WHEN** a protected request returns 401
- **THEN** the miniapp SHALL clear its session and offer login without showing another user's cached data

### Requirement: Demo boundary
The miniapp SHALL distinguish local calendar and simplified personal-day demos from authenticated real chart results.

#### Scenario: Calendar demo
- **WHEN** a user views the prepacked calendar or activity picker
- **THEN** the page SHALL identify the static date range and SHALL NOT present fixed-chart personal output as the logged-in user's analysis
