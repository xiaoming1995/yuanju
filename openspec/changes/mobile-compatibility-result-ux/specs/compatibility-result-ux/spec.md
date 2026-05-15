## ADDED Requirements

### Requirement: Mobile conclusion-first result hierarchy
The compatibility result page MUST present relationship conclusion content before dense professional chart details on mobile.

#### Scenario: User opens a compatibility result on a phone
- **WHEN** the result detail has loaded
- **THEN** the first major content block contains the participant names, compatibility level, summary, tags, and the primary report action when relevant
- **AND** dense participant chart snapshots appear after the quick result overview

### Requirement: Scannable dimension scores
The compatibility result page MUST render four dimension scores in a compact, scannable format suitable for mobile.

#### Scenario: User reviews dimension scores
- **WHEN** the page displays attraction, stability, communication, and practicality scores
- **THEN** each score includes a text label, numeric value, and visual strength indicator
- **AND** the layout does not require horizontal scrolling on mobile widths

### Requirement: Progressive professional detail access
The compatibility result page MUST make professional chart details available without forcing them into the first mobile viewport.

#### Scenario: User wants advanced chart details
- **WHEN** the user reaches the professional detail section
- **THEN** the page exposes both participants' four pillars and wuxing summaries
- **AND** detailed structural evidences remain available below the quick conclusion sections

### Requirement: Report insight fallback
The compatibility result page MUST show useful risk/advice content whether or not an AI compatibility report exists.

#### Scenario: Structured AI report exists
- **WHEN** the report contains structured risks and advice
- **THEN** the page presents those risks and advice as dedicated insight sections

#### Scenario: No structured AI report exists
- **WHEN** no structured report has been generated
- **THEN** the page derives key risk highlights from negative evidences when available
- **AND** the page clearly offers report generation for deeper guidance

### Requirement: Bottom navigation safe area
The compatibility result page MUST reserve enough bottom space on mobile so final actions and content are not obscured by the bottom navigation.

#### Scenario: User scrolls to the bottom on mobile
- **WHEN** the mobile bottom navigation is visible
- **THEN** the final result content remains readable and tappable above the navigation
