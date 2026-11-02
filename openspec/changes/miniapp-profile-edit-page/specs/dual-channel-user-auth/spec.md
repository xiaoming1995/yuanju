## ADDED Requirements

### Requirement: Random default nickname for new WeChat accounts
The backend SHALL assign a default display nickname consisting of “缘友” followed by six random decimal digits when creating a new WeChat miniapp account.

#### Scenario: New registration
- **WHEN** a valid unbound WeChat identity creates an account
- **THEN** the generated nickname SHALL be persisted with the new user and returned by login and profile APIs

#### Scenario: Returning or edited account
- **WHEN** an existing WeChat identity logs in again
- **THEN** its stored nickname SHALL be retained, including any user edits
- **THEN** existing accounts SHALL NOT be renamed by this change
