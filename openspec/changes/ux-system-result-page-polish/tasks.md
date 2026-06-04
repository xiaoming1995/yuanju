## 1. Baseline And Tests

- [x] 1.1 Record current frontend UX baseline metrics in `docs/superpowers/audits/2026-06-01-frontend-ux-baseline.md`
- [x] 1.2 Add a screenshot QA checklist for result page desktop and mobile validation
- [x] 1.3 Add source-level tests for UI primitive exports and required class hooks
- [x] 1.4 Add source-level tests for result page hero, action bar, segmented navigation, and mobile CTA hooks

## 2. Frontend UI Foundation

- [x] 2.1 Extend `frontend/src/index.css` with semantic color, spacing, typography, radius, and status tokens
- [x] 2.2 Create `PageShell` and `SectionPanel` primitives with CSS variable based layout styles
- [x] 2.3 Create shared `Button` primitive with primary, secondary, ghost, danger, disabled, and loading states
- [x] 2.4 Create shared `SegmentedTabs`, `StatusBadge`, and `EmptyState` primitives
- [x] 2.5 Create shared `ConfirmDialog` and `Toast` feedback primitives without adding a UI library
- [x] 2.6 Create shared `FormField` primitive for label, hint, error, and control composition
- [x] 2.7 Verify primitives do not depend on bazi, compatibility, admin, or auth business state

## 3. Result Page Decision-First Hero

- [x] 3.1 Extract a `ResultHeroSummary` component that shows chart identity, four-pillar overview, and deterministic core summary
- [x] 3.2 Extract a `ResultActionBar` component for AI interpretation, past events, and export actions
- [x] 3.3 Render the decision-first hero before professional detail sections in `ResultPage.tsx`
- [x] 3.4 Add mobile bottom action behavior for the primary AI interpretation action while respecting BottomNav and safe-area spacing
- [x] 3.5 Ensure missing AI report state still shows deterministic summary and generate-AI action

## 4. Result Page Segmented Structure

- [x] 4.1 Group existing result page detail content into overview, chart details, useful-god analysis, major luck, and AI interpretation sections
- [x] 4.2 Add segmented navigation with stable section ids and no page-level horizontal overflow on mobile
- [x] 4.3 Preserve existing data loading, saved history behavior, report generation behavior, and export behavior
- [x] 4.4 Keep professional modules available below the hero without compressing text below readable mobile sizes

## 5. Related Action Consistency

- [x] 5.1 Align history page card actions with result page labels for viewing result, past events, and export
- [x] 5.2 Align past-events page empty state with result/history routes so users can recover from missing chart context
- [x] 5.3 Ensure related actions use shared `Button` or compatible shared button styling

## 6. Verification

- [x] 6.1 Run the new source-level tests for UI primitives and result page structure
- [x] 6.2 Run `cd frontend && npm run lint`
- [x] 6.3 Run `cd frontend && npm run build`
- [x] 6.4 Manually inspect the result page at 390px width for no horizontal overflow and reachable primary CTA
- [x] 6.5 Manually inspect the result page at 1440px width for readable hierarchy and non-sticky desktop action layout
- [x] 6.6 Update the audit baseline document with final before/after notes for this change
