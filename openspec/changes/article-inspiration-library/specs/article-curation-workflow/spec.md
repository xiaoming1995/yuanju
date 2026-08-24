## ADDED Requirements

### Requirement: Admin Article Management Surface
The system SHALL provide an authenticated admin "资讯管理" area with tabs for articles, categories, tags, keywords, and collection task logs.

#### Scenario: Admin opens article management
- **WHEN** an authenticated administrator opens the article management page
- **THEN** the frontend SHALL show tabs for articles, categories, tags, keywords, and task logs

#### Scenario: Non-admin opens article management
- **WHEN** a request without a valid Admin JWT calls an admin article management endpoint
- **THEN** the system SHALL reject the request with an authentication error

### Requirement: Custom Category Management
The system SHALL allow administrators to create, update, list, sort, enable, and disable article categories.

#### Scenario: Admin creates category
- **WHEN** an authenticated administrator submits a valid new category name
- **THEN** the system SHALL create a category available for article assignment and filtering

#### Scenario: Admin disables category
- **WHEN** an authenticated administrator disables a category
- **THEN** the category SHALL stop appearing as an active frontend filter
- **THEN** existing article assignments SHALL remain stored

### Requirement: Custom Tag Management
The system SHALL allow administrators to create, update, list, enable, and disable article tags.

#### Scenario: Admin creates tag
- **WHEN** an authenticated administrator submits a valid new tag name
- **THEN** the system SHALL create a tag available for article assignment and filtering

#### Scenario: Admin disables tag
- **WHEN** an authenticated administrator disables a tag
- **THEN** the tag SHALL stop appearing as an active frontend filter
- **THEN** existing article-tag links SHALL remain stored

### Requirement: Keyword Management
The system SHALL allow administrators to maintain a simple active/inactive keyword list for article collection.

#### Scenario: Admin adds keyword
- **WHEN** an authenticated administrator adds a keyword
- **THEN** the system SHALL persist the keyword for future manual or scheduled collection

#### Scenario: Admin disables keyword
- **WHEN** an authenticated administrator disables a keyword
- **THEN** future collection tasks SHALL skip that keyword

### Requirement: Article Status Lifecycle
The system SHALL manage articles through candidate, published, rejected, taken_down, and deleted statuses.

#### Scenario: Candidate article is published
- **WHEN** an authenticated administrator publishes a candidate article
- **THEN** the system SHALL change its status to published
- **THEN** authenticated users SHALL be able to see it in article list and detail responses

#### Scenario: Published article is taken down
- **WHEN** an authenticated administrator takes down a published article
- **THEN** the system SHALL change its status to taken_down
- **THEN** authenticated users SHALL no longer see it in article list or detail responses

#### Scenario: Candidate article is rejected
- **WHEN** an authenticated administrator rejects a candidate article
- **THEN** the system SHALL change its status to rejected
- **THEN** authenticated users SHALL NOT see it in article list or detail responses

#### Scenario: Article is deleted
- **WHEN** an authenticated administrator deletes an article from any non-deleted status
- **THEN** the system SHALL change its status to deleted
- **THEN** authenticated users SHALL NOT see it in article list or detail responses

### Requirement: Article Audit Trail
The system SHALL record administrator and timestamp audit events for publish, reject, takedown, and delete actions.

#### Scenario: Admin publishes article
- **WHEN** an authenticated administrator publishes an article
- **THEN** the system SHALL write an audit event containing the article ID, admin ID, action, prior status, new status, and timestamp

#### Scenario: Admin takes down article
- **WHEN** an authenticated administrator takes down an article
- **THEN** the system SHALL write an audit event containing the article ID, admin ID, action, prior status, new status, and timestamp

### Requirement: Batch Review Actions
The system SHALL support batch publish, reject, and delete actions for article candidates and managed articles.

#### Scenario: Admin batch publishes candidates
- **WHEN** an authenticated administrator selects multiple candidate articles and submits a publish batch action
- **THEN** the system SHALL publish each eligible article
- **THEN** the system SHALL write audit events for each changed article

#### Scenario: Admin batch rejects candidates
- **WHEN** an authenticated administrator selects multiple candidate articles and submits a reject batch action
- **THEN** the system SHALL reject each eligible article
- **THEN** the system SHALL write audit events for each changed article

#### Scenario: Admin batch deletes articles
- **WHEN** an authenticated administrator selects multiple articles and submits a delete batch action
- **THEN** the system SHALL mark each eligible article as deleted
- **THEN** the system SHALL write audit events for each changed article

### Requirement: Publish Without AI Confirmation
The admin UI SHALL allow publishing an article without AI analysis only after an explicit confirmation.

#### Scenario: Admin publishes article without AI analysis
- **WHEN** an authenticated administrator attempts to publish an article without AI summary or writing breakdown
- **THEN** the frontend SHALL show a confirmation warning
- **THEN** the system SHALL publish the article only if the administrator confirms

### Requirement: Admin Takedown Handles Feedback
The system SHALL rely on admin takedown and audit records for source-owner or user feedback in V1.

#### Scenario: Admin receives external feedback
- **WHEN** an administrator decides an article should no longer be visible due to source-owner or user feedback
- **THEN** the administrator SHALL be able to take down the article
- **THEN** the system SHALL preserve an audit event for the takedown
