## ADDED Requirements

### Requirement: Published historical references are associated with a Ming Ge
The system SHALL store historical-reference entries separately from the generic celebrity library. Each entry MUST include a Ming Ge, figure name, era or identity, concise historical-memory text, at least one source title and URL, review status, and display order.

#### Scenario: A published reference is retrieved for a detected Ming Ge
- **WHEN** the public historical-reference API is requested with a Ming Ge that has published entries
- **THEN** it returns at most two entries for that exact Ming Ge in display order, including each entry's name, identity, historical-memory text, and source metadata

#### Scenario: Draft or archived references remain private
- **WHEN** the public historical-reference API is requested for a Ming Ge whose entries are draft, reviewed, or archived
- **THEN** it MUST NOT return those entries

### Requirement: Dayun interpretation is gated by verified source data
The system SHALL expose an entry's Dayun interpretation only when the entry is published and has exact-hour birth precision, a Bazi verification note, a source, a turning point, and a complete Dayun explanation. The explanation MUST describe correspondence with the turning point and MUST NOT state that a Dayun alone caused success.

#### Scenario: Verified Dayun information is displayed
- **WHEN** a published historical reference meets every Dayun display condition
- **THEN** the public API returns its Dayun period, turning point, and contextual explanation

#### Scenario: Birth data or source is insufficient
- **WHEN** a historical reference lacks any required Dayun display condition
- **THEN** the public API omits the Dayun fields while preserving any eligible historical-memory content

### Requirement: Historical-reference content is managed through review status
The administrator interface SHALL allow authorized administrators to create, edit, archive, and change the review status of a historical reference, including its sources, birth-data precision, turning point, and optional Dayun data.

#### Scenario: Administrator publishes a complete historical reference
- **WHEN** an administrator marks a complete reference as published
- **THEN** it becomes available from the public API for its exact Ming Ge

#### Scenario: Administrator attempts to publish invalid Dayun data
- **WHEN** an administrator enables Dayun display without all required verification fields
- **THEN** the system rejects the update and identifies that the Dayun verification data is incomplete

### Requirement: Result pages render historical references as Ming Ge context
The result page SHALL request published historical references only when the chart contains `ming_ge`. It SHALL render an unframed “古人映照” section only when at least one reference is returned, and it MUST present the content as Ming Ge context rather than a user-to-figure match.

#### Scenario: Ordinary mode renders concise historical context
- **WHEN** a chart with `ming_ge` is viewed in ordinary mode and the API returns references
- **THEN** the page displays each figure's name, identity, and historical-memory text, plus a boundary statement that the content is not a personal life-template or prediction

#### Scenario: Professional mode renders qualified details
- **WHEN** the same chart is viewed in professional mode
- **THEN** the page additionally displays the source, turning point, and only the Dayun information returned by the API

#### Scenario: A chart has no Ming Ge or no published references
- **WHEN** the chart lacks `ming_ge` or the reference API returns no entries or fails
- **THEN** the page omits the historical-reference section and continues rendering the rest of the result without an error

### Requirement: Historical references are not generated or matched at read time
The system MUST NOT use the LLM to invent historical figures, historical events, birth data, Dayun claims, or user similarity scores while serving a result page or calculating a chart.

#### Scenario: A result page is opened for an unsupported Ming Ge
- **WHEN** no published reference exists for the chart's Ming Ge
- **THEN** the system returns no historical references and does not substitute AI-generated content

### Requirement: Historical chart snapshots retain a usable Ming Ge
The system SHALL backfill `ming_ge` and `ming_ge_desc` when a historical chart snapshot contains the required Bazi fields but either Ming Ge field is absent. The backfill MUST preserve the remaining saved calculation result and persist the completed snapshot for subsequent reads.

#### Scenario: An old snapshot lacks Ming Ge fields
- **WHEN** an authenticated user opens a historical chart whose saved result has no `ming_ge` or `ming_ge_desc`
- **THEN** the system derives both fields from the saved Bazi result, returns them with the history response, and saves the completed snapshot

#### Scenario: A complete snapshot is loaded
- **WHEN** an authenticated user opens a historical chart whose saved result already has both Ming Ge fields
- **THEN** the system returns the saved Ming Ge fields without recalculating unrelated Bazi result fields
