## Context

User-facing pages are rendered below a fixed 64px Navbar and, on mobile, above a fixed BottomNav. Most pages use the global `.page` class, but `/profile` used `page-container`, and several pages have copied page-specific bottom padding to avoid mobile clipping.

The repeated regressions show that navigation avoidance belongs in the shared page shell rather than in individual pages.

## Goals / Non-Goals

**Goals:**
- Make `.page` the shared contract for ordinary user-facing pages.
- Reserve enough top space for the fixed Navbar on desktop and mobile.
- Reserve enough bottom space for the mobile BottomNav and safe-area inset.
- Keep existing visual spacing stable where page-specific styles already rely on it.
- Add static regression tests covering the shared rule and `/profile` adoption.

**Non-Goals:**
- Replace routing with a new React layout component in this change.
- Redesign Navbar or BottomNav.
- Change admin pages.
- Change business logic, APIs, or report generation behavior.

## Decisions

1. **Use CSS shell first, not a React wrapper.**
   - Rationale: The current app already has a `.page` class on most routes, so global CSS is the smallest reliable fix.
   - Alternative considered: introduce an `AppShell` component and wrap every route. That is cleaner long term but higher risk while mobile UX work is still moving quickly.

2. **Normalize `/profile` to `.page`.**
   - Rationale: Profile was the only ordinary user page using `page-container`, which bypassed the global top padding.
   - Alternative considered: keep `page-container` and duplicate shell styles. That preserves the inconsistency that caused the bug.

3. **Keep page-specific bottom padding only when it exceeds shell defaults.**
   - Rationale: Some result/history pages already reserve larger mobile space for dense actions and visual breathing room. The global shell provides a safe baseline; page overrides can remain intentional.

4. **Test CSS contracts with lightweight static tests.**
   - Rationale: Existing frontend tests are file-based and fast. They already caught these regressions without needing a full browser harness.

## Risks / Trade-offs

- **Risk:** Increasing global `.page` bottom padding may add extra whitespace on short pages.
  - **Mitigation:** Use conservative desktop baseline and preserve page-specific mobile overrides where needed.

- **Risk:** Some page-specific CSS may duplicate global shell spacing.
  - **Mitigation:** Do not remove all duplicates in this pass; only normalize the missing shell path and add tests first.

- **Risk:** Static tests cannot catch every visual overlap.
  - **Mitigation:** Pair tests with browser checks for `/profile`, `/history`, and a result page at mobile width.
