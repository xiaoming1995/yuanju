## Why

Mobile layout regressions have appeared repeatedly across `/profile`, result, history, and compatibility pages. A fixed QA matrix turns these ad hoc checks into a repeatable contract before more UX work continues.

## What Changes

- Add a maintained mobile page QA matrix for core user-facing routes.
- Add static tests that verify each matrix route exists, uses Navbar/BottomNav, and renders through the shared `.page` shell.
- Cover key mobile viewport sizes and layout expectations in one reusable test file.
- Keep this change test-focused; no business logic changes.

## Capabilities

### New Capabilities
- `mobile-page-qa-matrix`: Core user-facing pages have a repeatable mobile QA checklist and static route/layout coverage.

### Modified Capabilities
- None.

## Impact

- Affected files: frontend tests and OpenSpec change artifacts.
- No backend, database, API, or runtime dependency changes.
