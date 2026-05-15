## 1. Shell Contract Tests

- [x] 1.1 Add regression tests for global `.page` top and bottom navigation safe areas.
- [x] 1.2 Add regression tests that Profile page uses `.page` in all render states.

## 2. Shell Implementation

- [x] 2.1 Update global `.page` CSS to reserve fixed Navbar and BottomNav safe areas consistently.
- [x] 2.2 Normalize Profile page class names from `page-container` to `.page`.
- [x] 2.3 Remove Profile-only top/bottom shell padding that is covered by the shared shell.

## 3. Verification

- [x] 3.1 Verify `/profile`, `/history`, and a result page in mobile browser dimensions.
- [x] 3.2 Run frontend lint, static tests, build, and diff whitespace checks.
