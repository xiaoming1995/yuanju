## ADDED Requirements

### Requirement: Core mobile routes are listed in a QA matrix
The frontend test suite SHALL define a maintained QA matrix for core user-facing mobile routes.

#### Scenario: Route matrix covers core pages
- **WHEN** the mobile page QA matrix is loaded
- **THEN** it lists home, profile, history, compatibility, compatibility history, bazi result, and compatibility result routes

### Requirement: Matrix routes keep shared navigation shell
Every route in the mobile page QA matrix SHALL render with Navbar, BottomNav, and a page component using the shared `.page` shell.

#### Scenario: Static route shell validation
- **WHEN** static tests inspect the route and page source
- **THEN** each matrix route has Navbar and BottomNav in `App.tsx` and a page shell class in its page component

### Requirement: Mobile QA viewports are explicit
The QA matrix SHALL define the mobile viewport sizes used for manual or browser-based layout checks.

#### Scenario: Viewport checklist exists
- **WHEN** a tester reads the mobile QA matrix
- **THEN** the matrix includes 390x844, 375x812, and 360x740 viewport sizes
