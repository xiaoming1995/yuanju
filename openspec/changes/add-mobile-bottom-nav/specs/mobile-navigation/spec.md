## ADDED Requirements

### Requirement: Global Bottom Navigation
The system SHALL display a fixed bottom navigation bar exclusively for mobile devices.

#### Scenario: Mobile View
- **WHEN** the viewport width is less than or equal to 640px
- **THEN** the system displays a fixed bar at the bottom of the screen containing navigation icons.
- **AND** the page content has sufficient padding to not be obscured.

#### Scenario: Desktop View
- **WHEN** the viewport width is greater than 640px
- **THEN** the bottom navigation bar is hidden from view.

### Requirement: Navigation Items Rendering
The system SHALL render navigation items based on the user's authentication state.

#### Scenario: Unauthenticated User
- **WHEN** a user is not logged in 
- **THEN** the bottom navigation bar displays an item for "测算" (Home) and an item for "我的" (Login/Profile).

#### Scenario: Authenticated User
- **WHEN** a user is logged in
- **THEN** the bottom navigation bar displays "测算" (Home) and "历史" (History).
