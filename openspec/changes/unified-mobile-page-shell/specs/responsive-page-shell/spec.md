## ADDED Requirements

### Requirement: User pages reserve fixed navigation space
User-facing pages SHALL use a shared page shell that reserves space for the fixed top Navbar and mobile BottomNav.

#### Scenario: Top navigation does not cover first content
- **WHEN** a user opens a user-facing page with the fixed Navbar visible
- **THEN** the first page content starts below the Navbar safety area

#### Scenario: Mobile bottom navigation does not cover final content
- **WHEN** a user opens a user-facing page at a mobile viewport width with BottomNav visible
- **THEN** the final page content remains above the BottomNav and safe-area inset

### Requirement: Profile uses shared page shell
The profile page SHALL participate in the shared `.page` shell contract instead of using a separate page container convention.

#### Scenario: Profile layout follows shared shell
- **WHEN** the profile page renders loading, error, empty, or loaded states
- **THEN** the outer page element includes the shared page shell class
