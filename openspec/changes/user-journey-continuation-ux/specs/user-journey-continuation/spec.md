## ADDED Requirements

### Requirement: Guest bazi calculation SHALL continue after authentication
The system SHALL preserve an anonymous user's current bazi calculation context when the user chooses to log in or register from a result-oriented flow, and SHALL resume the intended authenticated action after successful authentication.

#### Scenario: Guest calculates a chart and logs in to generate a report
- **WHEN** an anonymous user completes a bazi calculation
- **AND** the user chooses to log in from the result page in order to generate an authenticated report
- **THEN** the frontend SHALL store the pending bazi journey with the original calculation input and intended report action
- **AND** after successful login the frontend SHALL create or resolve a user-owned chart for the authenticated user
- **AND** the user SHALL land back on the result flow for that authenticated chart rather than the generic homepage

#### Scenario: Guest registers from a result page
- **WHEN** an anonymous user completes a bazi calculation
- **AND** the user chooses to register from the result page
- **THEN** the registration flow SHALL preserve the pending bazi journey
- **AND** after successful registration the user SHALL return to the authenticated continuation target for that chart

#### Scenario: Authenticated chart recreation fails
- **WHEN** the user successfully authenticates with a pending bazi journey
- **AND** recreating the authenticated chart fails
- **THEN** the system SHALL keep the pending journey available for retry
- **AND** the user SHALL see a recoverable inline error or toast instead of being silently redirected to the homepage

#### Scenario: Pending journey is stale or invalid
- **WHEN** the user authenticates with a pending journey that is expired, malformed, or missing the original calculation input
- **THEN** the system SHALL discard the invalid pending journey
- **AND** the user SHALL continue to a normal authenticated fallback page

### Requirement: Post-auth redirects SHALL use safe local return targets
The login and registration flows SHALL support returning users to a safe local target after authentication while preventing open redirects and cross-role route confusion.

#### Scenario: Safe local next path is provided
- **WHEN** a user opens the login page with a valid local `next` target such as `/history` or `/result`
- **AND** authentication succeeds
- **THEN** the app SHALL navigate to that local path

#### Scenario: Unsafe next path is provided
- **WHEN** a user opens the login or registration page with an absolute URL, protocol-relative URL, malformed URL, or normal-user flow targeting an admin route
- **AND** authentication succeeds
- **THEN** the app SHALL ignore that target
- **AND** the app SHALL navigate to the pending journey target or normal authenticated fallback

#### Scenario: Register link preserves current safe target
- **WHEN** a user switches from login to registration while a safe `next` target exists
- **THEN** the registration link SHALL preserve that safe target
- **AND** successful registration SHALL resolve the same intended destination

### Requirement: Recent analysis continuation SHALL consider both bazi and compatibility records
The profile and archive entry points SHALL present a continuation path based on the user's most recent meaningful analysis across bazi charts and compatibility readings, not by a fixed feature priority.

#### Scenario: Latest item is a compatibility reading
- **WHEN** a user has both recent bazi charts and recent compatibility readings
- **AND** the most recent compatibility reading is newer than the most recent bazi chart
- **THEN** the primary continue action SHALL target the compatibility result

#### Scenario: Latest item is a bazi chart
- **WHEN** a user has both recent bazi charts and recent compatibility readings
- **AND** the most recent bazi chart is newer than the most recent compatibility reading
- **THEN** the primary continue action SHALL target the bazi result or its next relevant analysis action

#### Scenario: User has no saved analyses
- **WHEN** an authenticated user has no bazi chart and no compatibility reading
- **THEN** the profile or archive continuation area SHALL show a start-new-analysis action
- **AND** it SHALL NOT render an empty or broken continue link

### Requirement: Mobile navigation SHALL expose saved analyses as a primary destination
The mobile bottom navigation SHALL provide a clear primary destination for saved readings or analysis archives.

#### Scenario: Authenticated mobile user opens archive destination
- **WHEN** an authenticated mobile user taps the saved-reading or archive navigation item
- **THEN** the app SHALL navigate to a page where bazi history and compatibility history are reachable without first visiting profile

#### Scenario: Anonymous mobile user opens archive destination
- **WHEN** an anonymous mobile user taps the saved-reading or archive navigation item
- **THEN** the app SHALL navigate to login or registration with a safe local return target for the archive destination
- **AND** after authentication the user SHALL land on the archive destination

### Requirement: Result action feedback SHALL be consistent across user-facing result pages
User-facing result workflows SHALL use shared feedback patterns for generation, export/share, retry, and destructive actions.

#### Scenario: Export or share fails
- **WHEN** image export, PDF export, or share-card generation fails on a bazi or compatibility result page
- **THEN** the app SHALL show a non-blocking toast or inline error with a retry path
- **AND** it SHALL NOT use a native browser `alert()`

#### Scenario: User triggers a destructive action
- **WHEN** the user deletes a saved bazi chart or compatibility reading
- **THEN** the app SHALL use a shared confirmation dialog
- **AND** it SHALL NOT use a native browser `confirm()`

#### Scenario: Generation is in progress
- **WHEN** an AI report, polished report, past-events segment, or compatibility report is generating
- **THEN** the relevant action control SHALL show an in-flight state
- **AND** duplicate generation actions SHALL be disabled or ignored until the current request settles

#### Scenario: Generation fails after partial progress
- **WHEN** a streamed or long-running generation fails after the page has already shown progress
- **THEN** the page SHALL keep any usable generated content visible
- **AND** it SHALL show a clear retry affordance for the failed action

### Requirement: Homepage and profile SHALL emphasize active user intents over unavailable features
The homepage and profile page SHALL prioritize actionable analysis intents and SHALL avoid giving inactive or unavailable features the same visual weight as usable journeys.

#### Scenario: User lands on the homepage
- **WHEN** any user opens the homepage
- **THEN** the page SHALL make starting a bazi chart, continuing a saved analysis, and entering compatibility analysis visually discoverable as primary intents

#### Scenario: Precision-related input copy is displayed
- **WHEN** the page describes longitude or true-solar-time correction
- **THEN** the copy SHALL avoid implying city-level or astronomical precision unless the input actually captures that precision

#### Scenario: Profile contains coming-soon features
- **WHEN** the profile page includes disabled or coming-soon capabilities
- **THEN** they SHALL be visually secondary to saved analyses, recent continuation, history, compatibility, and settings actions
