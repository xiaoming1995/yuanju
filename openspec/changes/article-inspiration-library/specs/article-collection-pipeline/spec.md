## ADDED Requirements

### Requirement: Background Collection Boundary
The system SHALL run article collection as an independent background task, command, or scheduler path rather than inside ordinary frontend article requests.

#### Scenario: User opens article list
- **WHEN** an authenticated user requests the article list
- **THEN** the system SHALL read already stored published articles
- **THEN** the system SHALL NOT initiate Sogou WeChat collection during that request

#### Scenario: Collection task runs
- **WHEN** a manual or scheduled collection task starts
- **THEN** the task SHALL write candidate articles and task logs to the database outside ordinary user request flow

### Requirement: Sogou WeChat Search Provider
The system SHALL collect V1 article candidates from Sogou WeChat search using active admin keywords.

#### Scenario: Active keyword is collected
- **WHEN** a collection task runs with an active keyword
- **THEN** the system SHALL query the Sogou WeChat search provider for that keyword
- **THEN** the system SHALL transform discovered WeChat article search results into candidate article records

#### Scenario: Inactive keyword is skipped
- **WHEN** a collection task runs
- **AND** a keyword is inactive
- **THEN** the system SHALL NOT query the provider for that keyword

### Requirement: Collection Data Scope
The system SHALL persist article metadata, public search result information, and best-effort body content from collection.

#### Scenario: Search result is inserted
- **WHEN** the collection task stores a new candidate article
- **THEN** the system SHALL persist title, source name, original URL, cover URL if available, source publish time if available, and search snippet if available
- **AND** the system SHALL attempt to fetch cleaned article body content
- **AND** body fetch failure SHALL NOT block candidate insertion

### Requirement: URL Deduplication
The system SHALL deduplicate collected article candidates by normalized original URL.

#### Scenario: New original URL is found
- **WHEN** a collection task discovers an article URL that is not already stored
- **THEN** the system SHALL insert a candidate article

#### Scenario: Duplicate original URL is found
- **WHEN** a collection task discovers an article URL that is already stored
- **THEN** the system SHALL NOT insert another article row
- **THEN** the system SHALL record the result as a duplicate in the task counts

### Requirement: Manual Collection Trigger
The system SHALL allow authenticated administrators to manually trigger article collection.

#### Scenario: Admin manually starts collection
- **WHEN** an authenticated administrator triggers collection
- **THEN** the system SHALL create a collection task
- **THEN** the task SHALL collect using active keywords

#### Scenario: Non-admin manually starts collection
- **WHEN** a request without a valid Admin JWT attempts to trigger collection
- **THEN** the system SHALL reject the request with an authentication error

### Requirement: Scheduled Collection Configuration
The system SHALL allow administrators to configure global scheduled collection interval in minutes and maximum result count.

#### Scenario: Admin updates schedule config
- **WHEN** an authenticated administrator sets the schedule interval in minutes and sets a maximum collection count
- **THEN** the system SHALL persist the configuration
- **THEN** future scheduled collection SHALL use the persisted values

#### Scenario: Scheduler runs with config
- **WHEN** the scheduler starts a collection run
- **THEN** the task SHALL respect the configured interval and maximum result count

#### Scenario: Collection uses global max result count
- **WHEN** a collection run has multiple active keywords
- **THEN** the configured maximum result count SHALL apply to the whole task
- **AND** the task SHALL NOT collect up to the maximum once per keyword

### Requirement: Collection Task Logs
The system SHALL record collection task status, counts, timestamps, errors, and collected item details for admin inspection.

#### Scenario: Collection succeeds
- **WHEN** a collection task completes successfully
- **THEN** the system SHALL record started time, finished time, status, keyword count, found count, inserted count, duplicate count, and failed count

#### Scenario: Collection provider fails
- **WHEN** the Sogou provider fails for a keyword or task
- **THEN** the system SHALL record the failure status and error message in task logs
- **THEN** the admin task log view SHALL expose the failure reason

#### Scenario: Admin inspects collected task items
- **WHEN** an authenticated administrator opens a collection task detail list
- **THEN** the system SHALL show each collected item status, keyword, title when stored, source when stored, original link, and related article link when available
- **AND** failed items SHALL expose the item-level failure reason

### Requirement: Collection Retry
The system SHALL allow administrators to retry failed collection tasks or failed task items.

#### Scenario: Admin retries failed task
- **WHEN** an authenticated administrator retries a failed collection task
- **THEN** the system SHALL start a retry run for the failed work
- **THEN** existing article URLs SHALL still be deduplicated

#### Scenario: Admin retries failed item
- **WHEN** an authenticated administrator retries a failed collection item
- **THEN** the system SHALL reprocess that item
- **THEN** the system SHALL update the item status and error message based on the retry result
