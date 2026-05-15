## Why

Profile, history, result, and compatibility pages currently handle fixed Navbar and BottomNav spacing independently. This has already caused repeated regressions where mobile content is clipped by the bottom menu or the first card is hidden under the top navigation.

## What Changes

- Introduce a shared page-shell contract for user-facing pages that reserves top Navbar and bottom BottomNav safe areas.
- Normalize `/profile` to use the same `.page` shell semantics as other non-admin pages.
- Keep page-specific spacing only where a page needs extra visual rhythm, not for core navigation avoidance.
- Add regression tests for shared shell rules and key user pages.

## Capabilities

### New Capabilities
- `responsive-page-shell`: User-facing pages reserve fixed navigation safe areas consistently across desktop and mobile.

### Modified Capabilities
- None.

## Impact

- Affected frontend CSS: `frontend/src/index.css`, page-specific CSS files where duplicated navigation spacing can be simplified.
- Affected frontend pages: `ProfilePage` class usage and any user-facing route relying on `.page`.
- No API, database, or backend behavior changes.
