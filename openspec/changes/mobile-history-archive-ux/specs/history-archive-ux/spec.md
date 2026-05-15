## ADDED Requirements

### Requirement: Chart history archive presentation
The chart history page MUST present saved charts as an archive rather than a plain list.

#### Scenario: User opens chart history
- **WHEN** saved charts are loaded
- **THEN** the page shows archive-level context including record count and summary stats
- **AND** each chart card shows pillars, birth profile, gender, created date, and a clear view action

### Requirement: Archive cross-navigation
Archive pages MUST provide obvious navigation between chart records and compatibility records.

#### Scenario: User is viewing chart history
- **WHEN** the chart archive renders
- **THEN** the page shows an archive switcher linking to chart history and compatibility history

#### Scenario: User is viewing compatibility history
- **WHEN** the compatibility archive renders
- **THEN** the page shows an archive switcher linking to compatibility history and chart history

### Requirement: Compatibility history archive consistency
The compatibility history page MUST use the same archive visual language as chart history.

#### Scenario: User opens compatibility history
- **WHEN** compatibility readings are loaded
- **THEN** each record shows participants, level, tags, dimension scores, and a clear view action
- **AND** the page avoids inline layout styling for reusable archive UI

### Requirement: Archive empty states
Archive pages MUST provide direct next actions when no records exist.

#### Scenario: Chart archive has no records
- **WHEN** the user has no saved charts
- **THEN** the page offers an action to create a new chart

#### Scenario: Compatibility archive has no records
- **WHEN** the user has no compatibility readings
- **THEN** the page offers an action to create a compatibility reading

### Requirement: Mobile archive safe area
Archive pages MUST reserve enough bottom space so final cards and actions are not obscured by mobile bottom navigation.

#### Scenario: User scrolls to the bottom on mobile
- **WHEN** the mobile bottom navigation is visible
- **THEN** archive content remains readable and tappable above the navigation
