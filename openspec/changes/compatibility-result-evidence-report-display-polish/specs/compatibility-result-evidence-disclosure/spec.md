## ADDED Requirements

### Requirement: Key judgment evidence is visible without mandatory expansion
The compatibility result page SHALL show enough key judgment evidence by default for users to understand why a visible claim was made.

#### Scenario: First evidence claim is expanded by default
- **WHEN** the compatibility result has one or more claim-evidence links
- **THEN** the first key judgment evidence item is expanded or otherwise shows its reasoning and linked evidence without requiring a click

#### Scenario: Collapsed evidence still previews reasoning
- **WHEN** additional key judgment evidence items are collapsed
- **THEN** each collapsed item still shows the claim text and a concise reasoning preview

### Requirement: Evidence disclosure controls communicate state
The compatibility result page SHALL make evidence disclosure controls clear about whether evidence is hidden or visible.

#### Scenario: Evidence item can be collapsed after expansion
- **WHEN** a user expands an evidence item
- **THEN** the disclosure control communicates that the evidence can be hidden again

#### Scenario: Evidence item can be expanded from preview
- **WHEN** a user sees a collapsed evidence item
- **THEN** the disclosure control communicates that more evidence can be viewed

### Requirement: Professional details remain supportive rather than empty-looking
The compatibility result page SHALL show a compact professional-data summary even when dense chart details remain collapsible.

#### Scenario: Professional details are collapsed
- **WHEN** professional chart details are collapsed
- **THEN** the page still shows summary-level copy indicating that four pillars, five elements, and structured evidence are available

#### Scenario: Professional details are expanded
- **WHEN** professional chart details are expanded
- **THEN** the existing participant summaries and professional evidence groups remain available
