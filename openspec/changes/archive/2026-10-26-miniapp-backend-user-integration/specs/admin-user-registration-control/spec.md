## MODIFIED Requirements

### Requirement: Public Registration Setting
The system SHALL persist a site-level setting that controls whether ordinary users may self-register by email through the public registration endpoint. Miniapp first-login creation SHALL use a separate setting.

#### Scenario: Registration setting defaults enabled
- **WHEN** the system starts after the migration is applied and no explicit email registration setting exists
- **THEN** the system SHALL treat public email registration as enabled

#### Scenario: Admin reads registration setting
- **WHEN** an authenticated administrator requests the registration setting
- **THEN** the system SHALL return whether public email registration is currently enabled

#### Scenario: Admin updates registration setting
- **WHEN** an authenticated administrator updates the registration setting to enabled or disabled
- **THEN** the system SHALL persist the new value
- **THEN** later public email registration attempts SHALL use the updated value without requiring a server restart
- **THEN** miniapp first-login creation SHALL remain governed by its separate setting

#### Scenario: Non-admin cannot update registration setting
- **WHEN** a request without a valid Admin JWT attempts to update the registration setting
- **THEN** the system SHALL reject the request with an authentication error

## ADDED Requirements

### Requirement: Administrator bootstrap is private
The system SHALL NOT expose a public HTTP endpoint that creates administrators.

#### Scenario: Anonymous admin registration
- **WHEN** an unauthenticated client requests `/api/admin/auth/register`
- **THEN** the system SHALL NOT create an administrator

#### Scenario: First administrator setup
- **WHEN** an operator runs the private initialization command with valid local database configuration
- **THEN** the command SHALL create the first administrator only when no administrator exists
