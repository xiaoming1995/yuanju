## ADDED Requirements

### Requirement: Authenticated Article Access
The system SHALL require ordinary user authentication for all frontend article list and detail access.

#### Scenario: Unauthenticated user opens article list
- **WHEN** a request without a valid ordinary user JWT calls the article list endpoint
- **THEN** the system SHALL reject the request with an authentication error

#### Scenario: Authenticated user opens article list
- **WHEN** a request with a valid ordinary user JWT calls the article list endpoint
- **THEN** the system SHALL return published article summaries

#### Scenario: Unauthenticated user opens article detail
- **WHEN** a request without a valid ordinary user JWT calls an article detail endpoint
- **THEN** the system SHALL reject the request with an authentication error

### Requirement: Article List Filtering And Sorting
The system SHALL let authenticated users browse published articles by custom category, custom tag, keyword query, and latest or hot sorting.

#### Scenario: User filters by category
- **WHEN** an authenticated user requests the article list with a category filter
- **THEN** the system SHALL return only published articles assigned to that category

#### Scenario: User filters by tag
- **WHEN** an authenticated user requests the article list with a tag filter
- **THEN** the system SHALL return only published articles linked to that tag

#### Scenario: User searches by keyword
- **WHEN** an authenticated user requests the article list with a keyword query
- **THEN** the system SHALL search published article title, source, summary, and search snippet fields

#### Scenario: User sorts by latest
- **WHEN** an authenticated user requests the article list sorted by latest
- **THEN** the system SHALL order published articles by publication/display recency descending

#### Scenario: User sorts by hot
- **WHEN** an authenticated user requests the article list sorted by hot
- **THEN** the system SHALL order published articles by detail view count descending

### Requirement: Article Detail Display
The system SHALL display article source metadata, summary, reading support, writing breakdown, original link access, and an AI-generated-content notice on article detail pages when those fields exist.

#### Scenario: Article has AI analysis
- **WHEN** an authenticated user opens a published article with AI analysis
- **THEN** the system SHALL return the article metadata, summary, reading support, related topics, title pattern, structure outline, expression style, and rewrite angles
- **THEN** the frontend SHALL show a notice that the AI summary and writing breakdown are generated from public search information and are for reference only

#### Scenario: Article has no AI analysis
- **WHEN** an authenticated user opens a published article without AI analysis
- **THEN** the system SHALL return the article metadata and original link
- **THEN** the frontend SHALL avoid showing empty AI analysis sections

#### Scenario: Article has original source
- **WHEN** an authenticated user opens a published article with an original WeChat URL
- **THEN** the frontend SHALL provide a reading-original action that opens the original source

### Requirement: View Count Tracking
The system SHALL increment a published article's view count when an authenticated user opens its detail page.

#### Scenario: User opens detail page
- **WHEN** an authenticated user successfully retrieves a published article detail
- **THEN** the system SHALL increment that article's view count

#### Scenario: User requests missing article
- **WHEN** an authenticated user requests a nonexistent or non-published article detail
- **THEN** the system SHALL NOT increment any article view count

### Requirement: Original Link Click Tracking
The system SHALL record user-level original-link clicks for authenticated users.

#### Scenario: User clicks original link
- **WHEN** an authenticated user triggers the original-link click tracking endpoint for a published article
- **THEN** the system SHALL record the user ID, article ID, and clicked timestamp
- **THEN** the system SHALL increment the article's original click count

#### Scenario: User clicks original link for hidden article
- **WHEN** an authenticated user triggers original-link click tracking for a non-published article
- **THEN** the system SHALL reject the request
- **THEN** the system SHALL NOT write a click log

### Requirement: Frontend Navigation Entry
The frontend SHALL expose a top-navigation "资讯" entry for authenticated article access.

#### Scenario: User sees top navigation
- **WHEN** the frontend renders the normal site navigation
- **THEN** the navigation SHALL include a "资讯" link that routes to the article list page

### Requirement: Stored Body Content Rendering
The frontend SHALL render stored article body content on detail pages when body content was collected successfully.

#### Scenario: Article detail has collected body content
- **WHEN** an authenticated user opens a published article detail
- **AND** the article has `full_text_authorized=true` and non-empty `body_content`
- **THEN** the frontend SHALL render the body content as in-site text
- **AND** the frontend SHALL still expose original-link access

#### Scenario: Article detail has no collected body content
- **WHEN** an authenticated user opens a published article detail
- **AND** the article has no stored body content
- **THEN** the frontend SHALL render stored metadata, summary, tags, AI analysis, and original-link access
