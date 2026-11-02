## ADDED Requirements

### Requirement: Dedicated miniapp profile editor
The miniapp SHALL provide a dedicated Personal Profile page for avatar and nickname editing, accessible from the signed-in account header in Mine.

#### Scenario: Open existing profile
- **WHEN** a signed-in user taps the account header in Mine
- **THEN** the miniapp SHALL navigate to Personal Profile and load the current avatar and nickname
- **THEN** Mine SHALL NOT contain an inline profile editing form

#### Scenario: Load failure or missing session
- **WHEN** profile loading fails or the session is unavailable
- **THEN** the editor SHALL show a retry or return-to-login action and SHALL NOT permit saving stale profile data

### Requirement: Automatic profile saving
The profile editor SHALL save avatar and nickname changes independently using the existing authenticated profile APIs without save or cancel buttons and remain open after success.

#### Scenario: Avatar or nickname update
- **WHEN** a user selects an avatar or a changed nickname passes WeChat review
- **THEN** only selected changes SHALL be submitted
- **THEN** the edited field SHALL show saving and saved status without automatically leaving the page
- **THEN** returning to Mine SHALL refresh its displayed profile

#### Scenario: Failed save
- **WHEN** saving a field fails
- **THEN** the editor SHALL preserve that field's selection and offer a retry without resubmitting the other field

#### Scenario: Nickname review fails or edits are pending
- **WHEN** a nickname fails review or review times out
- **THEN** the miniapp SHALL NOT submit that nickname and SHALL allow another input attempt
- **THEN** the draft nickname SHALL be retained even if the native component clears its value or misreports a service timeout as a failed review
- **WHEN** the user attempts to leave with pending or failed edits
- **THEN** the miniapp SHALL request a platform leave warning where supported

### Requirement: First registration uses the profile page
The miniapp SHALL reuse Personal Profile for optional profile selection immediately after new WeChat registration.

#### Scenario: Complete an initial profile field
- **WHEN** a newly registered user selects an avatar or the changed nickname passes WeChat review
- **THEN** the miniapp SHALL automatically save that field without requiring the other field

#### Scenario: Skip initial profile
- **WHEN** a newly registered user skips the profile page
- **THEN** the miniapp SHALL return to Mine while retaining a usable logged-in account
- **THEN** later logins SHALL NOT automatically reopen profile editing
