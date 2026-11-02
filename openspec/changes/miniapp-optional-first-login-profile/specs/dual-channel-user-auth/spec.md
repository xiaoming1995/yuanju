## MODIFIED Requirements

### Requirement: Optional email compatibility
The system SHALL preserve email login for email users and return a stable user response for email-less users.

#### Scenario: Returning miniapp user signs in without profile details
- **WHEN** an existing miniapp user signs in with a default nickname or no avatar
- **THEN** Mine SHALL show the signed-in account with default profile details without automatically opening profile editing
- **THEN** profile setup SHALL NOT prevent login or chart use

#### Scenario: Miniapp user chooses WeChat profile details
- **WHEN** a signed-in miniapp user actively opens profile editing and selects an avatar or enters a nickname
- **THEN** the backend SHALL save each submitted detail to that ordinary user and return it on later profile requests
- **THEN** the user SHALL be able to cancel editing without changing the saved profile

#### Scenario: No email user profile
- **WHEN** an email-less miniapp user requests identity or profile data
- **THEN** the response SHALL contain an empty email and usable nickname and user ID

#### Scenario: Admin inspects miniapp user
- **WHEN** an administrator lists or searches users
- **THEN** email-less miniapp users SHALL be visible without scan errors and password reset SHALL not be offered for them
