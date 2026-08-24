## ADDED Requirements

### Requirement: Independent Article AI Configuration
The system SHALL provide article-specific AI provider and prompt configuration that is independent from Bazi report AI configuration.

#### Scenario: Admin configures article AI provider
- **WHEN** an authenticated administrator updates the article AI provider configuration
- **THEN** the system SHALL persist the article AI configuration separately from the active Bazi report provider

#### Scenario: Article analysis runs
- **WHEN** the system generates article AI analysis
- **THEN** it SHALL use the article AI provider and article prompt configuration
- **THEN** it SHALL NOT implicitly use the active Bazi report provider unless that provider is explicitly configured for article AI

### Requirement: Manual AI Analysis Generation
The system SHALL generate article AI analysis only when an authenticated administrator manually requests it for a candidate or managed article.

#### Scenario: Admin requests AI analysis
- **WHEN** an authenticated administrator requests AI analysis for an article
- **THEN** the system SHALL call the configured article AI provider
- **THEN** the system SHALL persist the generated analysis on the article

#### Scenario: Article is collected
- **WHEN** a collection task inserts a candidate article
- **THEN** the system SHALL NOT automatically generate AI analysis for that article

### Requirement: AI Input Scope
The system SHALL build article AI prompts from stored body content when available, plus search metadata and admin taxonomy hints.

#### Scenario: AI prompt is built
- **WHEN** the system prepares an article AI analysis request
- **THEN** the prompt SHALL include title, source name, source publish time if present, search snippet or public summary, original URL, and existing category/tag hints
- **AND** when stored body content is available and authorized, the prompt SHALL include a bounded body text excerpt

### Requirement: AI Analysis Output
The system SHALL store article AI analysis as structured fields for reading support and writing breakdown.

#### Scenario: AI generation succeeds
- **WHEN** article AI analysis completes successfully
- **THEN** the system SHALL store a one-sentence summary, key points, target readers, related topics, suggested tags, title pattern, opening style, structure outline, expression style, and rewrite angles when provided by the model

#### Scenario: AI suggests tags
- **WHEN** article AI analysis includes suggested tags
- **THEN** the system SHALL store the suggestions with the article analysis
- **THEN** the system SHALL NOT silently create active custom taxonomy tags without administrator action

### Requirement: AI Analysis Retry
The system SHALL let administrators retry failed or unsatisfactory article AI analysis for single articles or selected articles.

#### Scenario: AI generation fails
- **WHEN** article AI generation fails
- **THEN** the system SHALL record the failure status and error message for the article or analysis attempt

#### Scenario: Admin retries AI analysis
- **WHEN** an authenticated administrator retries AI analysis for an article
- **THEN** the system SHALL call the article AI provider again
- **THEN** the system SHALL replace or update the stored analysis with the successful retry result

### Requirement: AI Notice For Users
The frontend SHALL clearly notify users that article summaries and writing breakdowns are AI-generated from public search information and are for reference only.

#### Scenario: User views AI-analyzed article
- **WHEN** an authenticated user opens an article detail page that contains AI analysis
- **THEN** the frontend SHALL display a clear notice that the AI summary and writing breakdown are generated from public search information and are for reference only
