## ADDED Requirements

### Requirement: Persistent chart query count per user
The system SHALL maintain an independent cumulative count of successfully persisted bazi chart queries for each authenticated user. Repeated queries SHALL count individually, and chart deletion or activity event cleanup SHALL NOT reduce the count.

#### Scenario: Authenticated user queries the same chart twice
- **WHEN** a signed-in user successfully creates two chart records, including identical birth details
- **THEN** that user's cumulative query count increases by two

#### Scenario: Chart is deleted or creation fails
- **WHEN** a saved chart is deleted or chart creation transaction fails
- **THEN** deletion does not lower the cumulative count and failed creation returns an error without increasing it

#### Scenario: Guest queries a chart
- **WHEN** an unauthenticated guest creates a chart
- **THEN** no user's cumulative query count changes

#### Scenario: Miniapp account merges into web account
- **WHEN** a miniapp user is merged into a web user
- **THEN** the surviving user's cumulative count equals the sum of both users' prior counts

### Requirement: Admin query count presentation
The administrator user list SHALL show each user's cumulative chart query count. User detail SHALL show that cumulative count separately from the number of currently saved chart records. Existing API chart count fields SHALL remain compatible.

#### Scenario: Administrator opens user list and detail
- **WHEN** an authorized administrator views a user who previously deleted a chart
- **THEN** the list shows the cumulative count and detail labels both the cumulative count and the lower current saved chart count by their meaning

#### Scenario: Existing users migrate
- **WHEN** the new counter is initialized for existing users
- **THEN** it starts from the number of their currently saved chart records and does not claim to recover previously deleted charts
