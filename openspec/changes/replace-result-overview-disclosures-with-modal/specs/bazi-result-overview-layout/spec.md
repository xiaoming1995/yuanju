## ADDED Requirements

### Requirement: Overview explanations open in a contextual modal
The Bazi result overview SHALL use one contextual modal for full grade guidance, road guidance, vehicle evidence, and road evidence. The overview SHALL NOT render those full contents in page-level expandable disclosure blocks.

#### Scenario: User requests grade guidance
- **WHEN** a user activates the grade-guidance entry in the vehicle summary
- **THEN** the system SHALL open the modal with every supported grade and highlight the result's current grade

#### Scenario: User requests road guidance
- **WHEN** a user activates the road-guidance entry in the current-road summary
- **THEN** the system SHALL open the modal with the road-type explanations and highlight the current road when one is available

#### Scenario: Professional user requests algorithm evidence
- **WHEN** professional mode is active and vehicle or road evidence exists
- **THEN** the corresponding summary SHALL expose an evidence entry that opens the modal with the existing evidence labels, scores, and details

#### Scenario: No professional evidence is available
- **WHEN** a result has no evidence for a summary or professional mode is inactive
- **THEN** the system SHALL omit that evidence entry without leaving an empty support section

### Requirement: Explanation modal supports accessible dismissal and reading
The explanation modal SHALL identify its content with dialog semantics, keep its heading and close control available, and allow content to scroll independently of the result page.

#### Scenario: User closes the modal
- **WHEN** the user activates the close control, clicks the overlay, or presses `Escape`
- **THEN** the system SHALL close the modal and return focus to the entry that opened it

#### Scenario: User views the modal on a narrow viewport
- **WHEN** the viewport is at the mobile breakpoint
- **THEN** the modal SHALL remain within the viewport, retain an accessible close control, and allow long explanation content to scroll

## MODIFIED Requirements

### Requirement: Summary cards limit first-screen explanatory density
The primary vehicle summary SHALL show its current grade, vehicle type, score, concise positioning text, relevant summary tags, and compact entries for its supporting explanation. The current-road summary SHALL show its current road context and compact entries for its supporting explanation. Full grade-scale guidance, road metaphor guidance, and professional evidence SHALL be disclosed in the contextual modal outside the primary summary content.

#### Scenario: User views a vehicle result initially
- **WHEN** the result overview first renders
- **THEN** it SHALL show the current grade explanation and compact explanation entries without rendering the complete S-to-D guide or evidence list in the page flow

#### Scenario: User requests full grade guidance
- **WHEN** the user activates the grade-guidance entry
- **THEN** the system SHALL display every supported grade and its explanation in the contextual modal, highlighting the current grade

#### Scenario: Professional user requests evidence
- **WHEN** professional reading mode is active and the user activates a vehicle or road evidence entry
- **THEN** the system SHALL display the existing vehicle or road evidence in the contextual modal without changing its labels, scores, or details
