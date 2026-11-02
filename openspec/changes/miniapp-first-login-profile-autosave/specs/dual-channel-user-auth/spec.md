## ADDED Requirements

### Requirement: Optional profile capture at first miniapp registration
The system SHALL identify a newly created WeChat miniapp account and offer a skippable avatar and nickname selection immediately after its first successful login.

#### Scenario: New WeChat account
- **WHEN** a valid WeChat code creates a new miniapp account
- **THEN** the login response SHALL identify it as newly created
- **THEN** Mine SHALL present avatar and nickname selection after login without blocking account use

#### Scenario: Returning WeChat account
- **WHEN** a valid WeChat code resolves to an existing user
- **THEN** the login response SHALL identify it as an existing account
- **THEN** Mine SHALL NOT automatically show profile selection

#### Scenario: User completes both selections
- **WHEN** the new user selects an avatar and the chosen nickname passes WeChat review
- **THEN** the miniapp SHALL automatically save both using the authenticated profile endpoints without an additional save action
- **THEN** Mine SHALL show the saved avatar and nickname

#### Scenario: User skips or save fails
- **WHEN** the user cancels profile selection
- **THEN** login SHALL remain successful and default profile details SHALL remain usable
- **WHEN** saving either detail fails
- **THEN** the miniapp SHALL show the failure and retain the selections for retry
