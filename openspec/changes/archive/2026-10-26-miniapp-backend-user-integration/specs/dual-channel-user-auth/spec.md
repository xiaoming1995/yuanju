## ADDED Requirements

### Requirement: Shared ordinary user identity
The system SHALL retain existing email users and their IDs while allowing a WeChat miniapp identity without an email or password to reference the same ordinary user model.

#### Scenario: Existing web user
- **WHEN** an existing email user registers or logs in through the web endpoints
- **THEN** the existing response and ordinary user JWT SHALL remain usable
- **THEN** existing charts SHALL remain owned by the same user ID

#### Scenario: New miniapp user
- **WHEN** a valid unbound WeChat code is exchanged while miniapp registration is enabled
- **THEN** the system SHALL create one ordinary user with no synthetic email or password
- **THEN** the system SHALL store a unique AppID/OpenID association and return an ordinary user JWT

### Requirement: Secure WeChat login and binding
The backend SHALL exchange WeChat login codes server-side and SHALL NOT disclose the AppSecret or session key to clients.

#### Scenario: Repeat login
- **WHEN** the same AppID/OpenID logs in again or two logins race
- **THEN** both SHALL resolve to the same user ID without duplicate users

#### Scenario: Existing web user binds
- **WHEN** a web user authenticated by ordinary JWT supplies a valid unbound WeChat code
- **THEN** that identity SHALL bind to the existing user ID and subsequent miniapp login SHALL see existing history

#### Scenario: Conflicting binding
- **WHEN** an identity already belongs to another user
- **THEN** binding SHALL return a conflict and SHALL NOT move any data

#### Scenario: Invalid code or disabled account
- **WHEN** WeChat exchange fails or the mapped user is disabled/deleted
- **THEN** login SHALL fail without a usable JWT

### Requirement: Optional email compatibility
The system SHALL preserve email login for email users and return a stable user response for email-less users.

#### Scenario: No email user profile
- **WHEN** an email-less miniapp user requests identity or profile data
- **THEN** the response SHALL contain an empty email and usable nickname and user ID

#### Scenario: Admin inspects miniapp user
- **WHEN** an administrator lists or searches users
- **THEN** email-less miniapp users SHALL be visible without scan errors and password reset SHALL not be offered for them
