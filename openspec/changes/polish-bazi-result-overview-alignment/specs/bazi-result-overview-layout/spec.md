## ADDED Requirements

### Requirement: Result overview presents the conclusion before supporting modules
The Bazi result page SHALL render the `命盘总评` conclusion before the vehicle and current-road summaries in both visual and DOM reading order.

#### Scenario: Result has vehicle and road data
- **WHEN** a result includes a vehicle profile and current dayun road
- **THEN** the user SHALL encounter `命盘总评` before `命盘座驾` and `当前路况`

#### Scenario: Assistive technology reads the overview
- **WHEN** the overview is read in DOM order
- **THEN** its sequence SHALL be conclusion, vehicle summary, current-road summary, and supporting explanation

### Requirement: Vehicle and road summaries remain visually independent
On desktop widths, the vehicle and current-road summary cards SHALL top-align and use their own content-driven heights. A longer vehicle summary SHALL NOT stretch the current-road card's background to an artificial matching height.

#### Scenario: Vehicle explanation is longer than road summary
- **WHEN** vehicle data includes formation and grade context while current-road data is short
- **THEN** both cards SHALL share a top edge and the road card SHALL end after its own content

#### Scenario: Narrow viewport
- **WHEN** the result page is displayed at the mobile breakpoint
- **THEN** the summaries SHALL render as one column with vehicle before current road

### Requirement: Summary cards limit first-screen explanatory density
The primary vehicle summary SHALL show its current grade, vehicle type, score, concise positioning text, and relevant summary tags. Full grade-scale guidance, road metaphor guidance, and professional evidence SHALL be progressively disclosed outside the primary summary content.

#### Scenario: User views a vehicle result initially
- **WHEN** the result overview first renders
- **THEN** it SHALL show the current grade explanation without rendering the complete S-to-D guide as an expanded list

#### Scenario: User requests full grade guidance
- **WHEN** the user expands the grade guidance control
- **THEN** the system SHALL display every supported grade and its explanation, highlighting the current grade

#### Scenario: Professional user requests evidence
- **WHEN** professional reading mode is active and the user expands the evidence control
- **THEN** the system SHALL display the existing vehicle or road evidence without changing its labels, scores, or details

### Requirement: Overview visual primitives are consistent
The conclusion, vehicle summary, and current-road summary SHALL use a consistent header hierarchy, badge treatment, horizontal padding, and spacing rhythm appropriate to their roles.

#### Scenario: Desktop overview is displayed
- **WHEN** the page renders at desktop width
- **THEN** summary-card headers and badges SHALL align to a common top rhythm and text SHALL remain within each card's bounds

#### Scenario: Result lacks vehicle-road data
- **WHEN** a result does not include vehicle or road data
- **THEN** the conclusion and remaining result sections SHALL render without an empty overview band or broken spacing
