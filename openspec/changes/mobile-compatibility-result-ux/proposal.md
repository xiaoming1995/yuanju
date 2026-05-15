## Why

The compatibility input flow is now mobile-friendly, but the result page still presents all content with similar visual weight. On phones, users need to understand the relationship conclusion, key scores, risks, and next action before scanning professional chart details.

## What Changes

- Rework the compatibility result page so mobile users see a conclusion-first layout.
- Present dimension scores as scannable mobile-friendly score rows with visual strength bars.
- Promote duration assessment, risks, and advice above dense chart/evidence detail on mobile.
- Move professional chart snapshots and detailed evidences into lower-priority, collapsible-style sections.
- Preserve desktop information density and the existing backend API contract.

## Capabilities

### New Capabilities
- `compatibility-result-ux`: Mobile-first presentation and hierarchy requirements for compatibility result pages.

### Modified Capabilities
- None.

## Impact

- Frontend result page: `frontend/src/pages/CompatibilityResultPage.tsx`
- Frontend styling: `frontend/src/pages/CompatibilityResultPage.css`
- Frontend static tests for result page mobile hierarchy
- No backend API, database, or algorithm changes
