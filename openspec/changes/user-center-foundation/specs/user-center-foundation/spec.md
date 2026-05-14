## ADDED Requirements

### Requirement: Authenticated Profile Overview
The system SHALL provide an authenticated user center overview for ordinary users.

#### Scenario: Fetch profile overview
- **WHEN** an authenticated user requests the profile overview
- **THEN** the system SHALL return the user's basic account information
- **THEN** the system SHALL return account-level statistics for bazi charts, AI reports, and compatibility readings
- **THEN** the system SHALL only include data owned by the authenticated user

#### Scenario: Reject unauthenticated profile access
- **WHEN** a request without a valid ordinary user JWT accesses the profile overview
- **THEN** the system SHALL reject the request with an authentication error

### Requirement: Recent Account Activity
The system SHALL include recent user activity summaries in the user center.

#### Scenario: Recent bazi chart activity
- **WHEN** an authenticated user opens the profile overview
- **THEN** the profile data SHALL include a bounded list of the user's most recent bazi chart summaries
- **THEN** each chart summary SHALL include enough information for the frontend to link to the existing chart detail or history flow

#### Scenario: Recent compatibility activity
- **WHEN** an authenticated user opens the profile overview and has compatibility readings
- **THEN** the profile data SHALL include a bounded list of recent compatibility reading summaries
- **THEN** each compatibility summary SHALL include enough information for the frontend to link to the existing compatibility detail flow

### Requirement: Profile Page Navigation
The frontend SHALL provide a dedicated personal center page for logged-in users.

#### Scenario: Logged-in user can enter profile page
- **WHEN** a user is authenticated
- **THEN** the top-level navigation SHALL provide access to the profile page
- **THEN** the mobile bottom navigation SHALL expose the profile page as the user's "我的" area

#### Scenario: Unauthenticated user is guided to login
- **WHEN** an unauthenticated user attempts to access the profile page
- **THEN** the frontend SHALL guide the user to login instead of showing private account data

### Requirement: Future Commercial Feature Entrypoints
The user center SHALL reserve clear, non-functional entrypoints for future recharge and PDF template customization features.

#### Scenario: Recharge entrypoint is not yet active
- **WHEN** the user center displays recharge or credit-related entrypoints before the recharge system exists
- **THEN** those entrypoints SHALL clearly indicate that the feature is not yet available
- **THEN** the system SHALL NOT create orders, change balances, or simulate successful payments

#### Scenario: PDF template entrypoint is not yet active
- **WHEN** the user center displays PDF template customization before template settings exist
- **THEN** the entrypoint SHALL clearly indicate that the feature is not yet available
- **THEN** the existing PDF export flow SHALL continue using the current default print layout
