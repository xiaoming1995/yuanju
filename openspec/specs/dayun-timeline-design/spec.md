# dayun-timeline-design Specification

## Purpose
TBD - created by archiving change replicate-dayun-timeline-design. Update Purpose after archive.
## Requirements
### Requirement: Desktop Dayun timeline matches approved mockup composition
The system SHALL render the desktop Dayun timeline as a single large mockup-aligned panel containing a title, metadata row, Dayun card strip, Liunian card panel, Dayun summary strip, and footer disclaimer.

#### Scenario: Desktop user views a calculated Bazi result
- **WHEN** the result page renders on a desktop viewport
- **THEN** the Dayun timeline area contains the title "大运时间轴", start-yun metadata, gender metadata, chart pillar metadata, a Dayun card strip, a Liunian panel, a Dayun summary strip, and a disclaimer line.

#### Scenario: Desktop panel avoids inherited generic card layout
- **WHEN** the Dayun timeline section renders on desktop
- **THEN** its visual container uses the timeline-specific border, background, padding, and internal hierarchy rather than a generic result-page card title plus nested child card.

### Requirement: Desktop Dayun card strip displays ten compact cards in one row
The system SHALL display ten compact Dayun cards in one row on desktop when sufficient Dayun display data is available.

#### Scenario: Desktop timeline has ten display periods
- **WHEN** the selected chart has ten displayable Dayun periods
- **THEN** the desktop Dayun strip renders ten cards across one row without horizontal document overflow.

#### Scenario: Current Dayun is highlighted
- **WHEN** the current year falls inside a Dayun period
- **THEN** that Dayun card shows a gold border, a "当前" badge, emphasized ganzhi text, and a bottom active indicator matching the reference mockup.

#### Scenario: Dayun card content matches mockup density
- **WHEN** a Dayun card renders
- **THEN** it displays period index, age range, ganzhi, ten-god label, and Gregorian year range without long explanatory paragraphs.

### Requirement: Desktop Liunian panel displays ten compact cards in one row
The system SHALL display the selected Dayun's Liunian years as ten compact cards in one row on desktop.

#### Scenario: Selected Dayun renders Liunian card strip
- **WHEN** a Dayun period is selected
- **THEN** the Liunian panel title shows "{干支}大运流年", the subtitle shows the age and Gregorian range, and ten Liunian cards render in one row.

#### Scenario: Current year marker matches reference
- **WHEN** a Liunian card represents the current year
- **THEN** it shows the current-year dot/badge treatment and gold active border shown in the reference design.

#### Scenario: Transition year marker matches reference
- **WHEN** a Liunian card represents a transition year
- **THEN** it shows a corner or ribbon marker for "交脱" with the transition date visible or accessible in the card.

#### Scenario: Focus year marker matches reference
- **WHEN** a Liunian card is marked as a focus year
- **THEN** it shows a compact "重点" badge in the same position and visual language as the mockup.

### Requirement: Dayun summary strip appears below Liunian panel
The system SHALL render a Dayun summary strip below the Liunian cards with concise summary text and compact tags.

#### Scenario: Summary strip renders on desktop
- **WHEN** the Liunian panel has rendered
- **THEN** a "大运总览" strip appears below it with a short paragraph and compact tags such as ten-god main qi, five-element main qi, and trend keywords.

#### Scenario: Summary strip remains compact
- **WHEN** summary content is long
- **THEN** the strip preserves a single-row or compact wrapped layout without pushing the timeline into a verbose report section.

### Requirement: Mobile Dayun timeline uses dedicated phone-style layout
The system SHALL render the mobile timeline using the phone-style compact layout from the reference image rather than a simple scaled desktop panel.

#### Scenario: Mobile user views Dayun timeline
- **WHEN** the result page renders on a mobile viewport
- **THEN** the timeline shows an app-like top bar, compact metadata, two-column Dayun cards, two-column Liunian cards, and a bottom Dayun summary control.

#### Scenario: Mobile cards use two-column layout
- **WHEN** the mobile viewport width is used
- **THEN** Dayun cards render in two columns and Liunian cards render in two columns without horizontal document overflow.

#### Scenario: Mobile active states remain visible
- **WHEN** a mobile card is current, transition, or focus
- **THEN** the corresponding marker remains visible within the compact card without covering the ganzhi or ten-god text.

### Requirement: Timeline preserves existing interactions
The system SHALL preserve existing Dayun and Liunian interactions while changing the visual structure.

#### Scenario: User selects a Dayun card
- **WHEN** the user clicks a Dayun card
- **THEN** the selected Dayun becomes active and the Liunian panel updates to that period.

#### Scenario: User opens Liuyue detail
- **WHEN** the user clicks a Liunian card
- **THEN** the existing Liuyue drawer opens for that year and ganzhi.

#### Scenario: User opens Shensha annotation
- **WHEN** a Shensha chip with an annotation is clicked
- **THEN** the existing Shensha annotation behavior remains available.

### Requirement: Timeline implementation has visual verification coverage
The system SHALL include automated checks and browser verification for desktop and mobile timeline fidelity.

#### Scenario: Static tests cover structure
- **WHEN** frontend static tests run
- **THEN** they verify the timeline-specific wrapper, ten-card desktop grids, mobile two-column grids, current/transition/focus markers, summary strip, and absence of verbose explanation cards.

#### Scenario: Browser verification covers runtime layout
- **WHEN** browser verification runs against the local app
- **THEN** it confirms desktop Dayun and Liunian rows, mobile two-column card grids, and no horizontal overflow.

