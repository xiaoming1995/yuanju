## ADDED Requirements

### Requirement: Admin opens one user's chart query history
The administrator user list SHALL link each user's name and cumulative chart query count to the existing chart records page scoped by that user's ID. The user detail SHALL provide the same link, while its existing detail action remains available.

#### Scenario: Click user with saved charts
- **WHEN** an administrator clicks a user in the user list
- **THEN** the chart page shows only that user's chart queries and identifies the scoped view

#### Scenario: Click cumulative chart query count
- **WHEN** an administrator clicks a user's cumulative chart query count, including zero
- **THEN** the chart page shows only that user's chart queries or its empty state

#### Scenario: Inspect a chart and return
- **WHEN** an administrator opens a chart from the scoped list and returns from chart detail
- **THEN** the list remains scoped to the selected user

#### Scenario: User has no email or any chart queries
- **WHEN** an administrator opens a miniapp user without an email or a user who never queried a chart
- **THEN** the ID filter still identifies the selected user and the page shows that user's records or a clear empty state

### Requirement: Exact admin chart filtering
The protected admin chart list API SHALL accept an optional user UUID and apply it to both total count and paginated rows. A scoped request SHALL return query history, including deleted entries. An absent user ID SHALL retain the existing full-site saved-chart behavior.

#### Scenario: Scoped request with dates
- **WHEN** an administrator requests chart records with a valid user ID and date range
- **THEN** only that user's records within the date range appear and pagination totals use the same filters

#### Scenario: Invalid user ID
- **WHEN** an administrator supplies an invalid user ID
- **THEN** the API returns a client error rather than silently broadening to all users

#### Scenario: Return to all charts
- **WHEN** an administrator chooses to leave the user scope
- **THEN** the chart page returns to the full-site list without the user ID filter

### Requirement: Deleted queries retain minimal audit history
Each authenticated chart query that is successfully persisted SHALL create one durable history entry with user, chart ID and query time. Deleting the chart SHALL retain only query time and deleted status for administrators, while removing birth data, pillars, reports and the user-visible chart. Account deletion SHALL remove its query history; account merge SHALL transfer it to the surviving user.

#### Scenario: User creates and deletes a chart
- **WHEN** an authenticated user successfully creates a chart and later deletes it
- **THEN** their admin query list still includes that query time and deleted state, without a working chart detail link or birth and chart content

#### Scenario: Account merge
- **WHEN** a miniapp user merges into an existing web account
- **THEN** prior query history belongs to the surviving account without duplicate entries

#### Scenario: Migration backfill
- **WHEN** the history table is introduced
- **THEN** current charts and any recoverable recent creation events become history entries without fabricating times for older deleted charts
