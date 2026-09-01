## ADDED Requirements

### Requirement: Result overview uses distinct primary module boundaries
The result page SHALL render the initial result overview as distinct identity, conclusion, and summary regions with visible but restrained boundaries. The page SHALL NOT rely solely on whitespace or a single internal divider to distinguish the identity region, the conclusion, the natal vehicle summary, and the current road summary.

#### Scenario: Desktop result overview renders complete data
- **WHEN** a calculated chart includes stem-level guidance, a vehicle profile, and a current dayun road
- **THEN** the page SHALL render an identity region, a conclusion region, and separate natal vehicle and current road summary panels in that reading order

#### Scenario: Result overview has incomplete optional data
- **WHEN** a calculated chart does not include a vehicle profile or a current dayun road
- **THEN** the remaining overview regions SHALL retain their boundaries and SHALL NOT render an empty placeholder panel

### Requirement: Identity region groups chart identity and reading controls
The result page SHALL group birth information, four pillars, stem-level yongshen guidance, ming ge label when present, and reading-mode controls within the identity region. The region SHALL visually prioritize the four pillars over supporting labels and controls.

#### Scenario: Chart has stem-level yongshen guidance
- **WHEN** the result returns primary, usable, and adverse stem guidance
- **THEN** the identity region SHALL show the three categories as compact, readable guidance rows without replacing their existing meanings

#### Scenario: User switches reading mode
- **WHEN** the user selects simple mode or professional mode from the identity region
- **THEN** the selected mode SHALL update using the existing reading-mode behavior and the identity region SHALL remain visually intact

### Requirement: Conclusion region presents the primary interpretation separately
The result page SHALL render the chart verdict in a dedicated conclusion region separate from the identity and summary regions. The conclusion region SHALL provide a visible route to the existing judgment evidence without requiring the user to scan the summary panels.

#### Scenario: User reads a calculated chart
- **WHEN** the result overview is displayed
- **THEN** the chart verdict and its supporting keyword summary SHALL appear before the natal vehicle and current road panels

#### Scenario: User requests judgment evidence
- **WHEN** the user activates the conclusion region's evidence action
- **THEN** the page SHALL navigate to the existing relevant evidence section according to the active reading mode

### Requirement: Natal vehicle and current road panels use a shared summary layout
The result page SHALL display the natal vehicle and current road as individually bounded panels with shared heading, status, spacing, and action conventions. On wide viewports the panels SHALL be arranged in a top-aligned two-column summary layout; on narrow viewports they SHALL appear in natal-vehicle then current-road order.

#### Scenario: Wide viewport shows both summaries
- **WHEN** the viewport supports a two-column overview and both summaries are available
- **THEN** the natal vehicle and current road panels SHALL be side by side with independently visible outer boundaries

#### Scenario: Narrow viewport shows both summaries
- **WHEN** the viewport width is within the mobile breakpoint
- **THEN** the natal vehicle panel SHALL appear before the current road panel without horizontal overflow or clipped text

### Requirement: Chart utility actions do not interrupt result reading
The result page SHALL place saved-chart naming and related chart utility actions in the identity region's auxiliary action area when those actions are available. The actions SHALL retain their existing availability conditions and behavior.

#### Scenario: Authenticated user opens a saved chart
- **WHEN** the result belongs to an accessible saved chart
- **THEN** the chart-name input, save action, and related utility actions SHALL be available adjacent to the identity region rather than between the summary and later reading sections

#### Scenario: Guest or unsaved result is displayed
- **WHEN** no saved chart utility actions are available
- **THEN** the identity region SHALL not reserve blank action space
