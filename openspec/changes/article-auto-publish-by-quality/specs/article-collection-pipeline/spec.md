## MODIFIED Requirements

### Requirement: Scheduled Collection Configuration
The system SHALL allow administrators to configure global scheduled collection interval in minutes, maximum result count, and conservative auto-publish rules.

#### Scenario: Admin updates schedule config
- **WHEN** an authenticated administrator sets the schedule interval in minutes and sets a maximum collection count
- **THEN** the system SHALL persist the configuration
- **AND** future scheduled collection SHALL use the persisted values

#### Scenario: Admin updates auto-publish config
- **WHEN** an authenticated administrator updates article auto-publish settings
- **THEN** the system SHALL persist whether auto-publish is enabled
- **AND** the system SHALL persist the minimum auto-publish quality score
- **AND** the system SHALL persist whether readable body content is required
- **AND** the system SHALL persist the maximum number of auto-published articles per run

#### Scenario: Scheduler runs with config
- **WHEN** the scheduler starts a collection run
- **THEN** the task SHALL respect the configured interval and maximum result count

#### Scenario: Collection uses global max result count
- **WHEN** a collection run has multiple active keywords
- **THEN** the configured maximum result count SHALL apply to the whole task
- **AND** the task SHALL NOT collect up to the maximum once per keyword

### Requirement: Collection Task Logs
The system SHALL record collection task status, counts, timestamps, errors, collected item details, and auto-publish outcomes for admin inspection.

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
- **AND** automatically published items SHALL expose that they were published automatically

## ADDED Requirements

### Requirement: Quality-Based Auto Publish
The system SHALL optionally publish newly collected article candidates automatically when conservative quality and body-content gates pass.

#### Scenario: Auto-publish is disabled
- **WHEN** a collection task inserts a new candidate article
- **AND** auto-publish is disabled
- **THEN** the article SHALL remain in candidate status regardless of quality score

#### Scenario: Candidate meets auto-publish gates
- **WHEN** a collection task inserts a new candidate article
- **AND** auto-publish is enabled
- **AND** the article quality score is greater than or equal to the configured auto-publish threshold
- **AND** the article has authorized non-empty body content
- **AND** the per-run auto-publish cap has not been reached
- **THEN** the system SHALL publish the article automatically
- **AND** authenticated users SHALL be able to see it in article list and detail responses

#### Scenario: Candidate score is below auto-publish threshold
- **WHEN** a collection task inserts a new candidate article
- **AND** auto-publish is enabled
- **AND** the article quality score is below the configured auto-publish threshold
- **THEN** the article SHALL remain in candidate status

#### Scenario: Candidate has no readable body content
- **WHEN** a collection task inserts a new candidate article
- **AND** auto-publish is enabled
- **AND** readable body content is required
- **AND** the article has no authorized non-empty body content
- **THEN** the article SHALL remain in candidate status

#### Scenario: Per-run cap is reached
- **WHEN** a collection task has already automatically published the configured maximum number of articles
- **AND** another newly inserted candidate meets the quality and body gates
- **THEN** the additional article SHALL remain in candidate status

#### Scenario: Duplicate article is encountered
- **WHEN** a collection task encounters an article URL that already exists
- **THEN** the system SHALL NOT auto-publish the existing article as part of duplicate handling
