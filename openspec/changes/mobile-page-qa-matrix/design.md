## Context

The app now has a shared `.page` shell for fixed Navbar and BottomNav spacing, but coverage is split across several narrow tests. The next useful guardrail is a route-level QA matrix that captures which pages must be checked for mobile layout regressions.

## Goals / Non-Goals

**Goals:**
- Define the core mobile QA pages in one place.
- Verify routes in the matrix are wired in `App.tsx`.
- Verify each route keeps Navbar and BottomNav in the user-facing route wrapper.
- Verify each page component participates in the shared page shell.
- Document the viewport sizes and expected checks used for manual or browser QA.

**Non-Goals:**
- Add Playwright or another browser testing dependency.
- Replace manual browser checks completely.
- Change page visual design or business behavior.

## Decisions

1. **Use a test-local matrix module.**
   - Rationale: The matrix is QA metadata, not runtime app behavior.
   - Alternative: Put it in `src/lib`; rejected because production code does not need to ship this metadata.

2. **Use static route/layout tests first.**
   - Rationale: Existing test style is fast Node-based file inspection and catches the specific regression class.
   - Alternative: Add browser automation to CI; useful later, but it requires dependency/runtime decisions outside this scope.

3. **Cover only core user routes.**
   - Rationale: Admin routes have a separate layout and mobile navigation model.

## Risks / Trade-offs

- **Risk:** Static tests cannot prove exact pixel layout.
  - **Mitigation:** The matrix includes viewport and check metadata for browser QA, while static tests prevent route and shell drift.

- **Risk:** A route can pass static checks and still have internal overflow.
  - **Mitigation:** Future browser tests can consume the same matrix without changing the route list.
