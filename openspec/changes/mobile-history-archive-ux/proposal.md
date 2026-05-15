## Why

History is currently split between a basic chart list, a separate compatibility list, and profile shortcuts. After improving input and result pages, the next retention bottleneck is helping returning users find saved charts and compatibility readings quickly on mobile.

## What Changes

- Upgrade `/history` into a mobile-friendly chart archive page.
- Add clear archive stats and cross-entry navigation between chart records and compatibility records.
- Make chart history cards more informative and action-oriented without changing backend data.
- Restyle `/compatibility/history` to match the archive visual language and remove inline layout styles.
- Preserve existing routes and API contracts.

## Capabilities

### New Capabilities
- `history-archive-ux`: Mobile-first archive presentation for saved chart and compatibility records.

### Modified Capabilities
- None.

## Impact

- Frontend chart history page: `frontend/src/pages/HistoryPage.tsx`
- Frontend chart history styles: `frontend/src/pages/HistoryPage.css`
- Frontend compatibility history page: `frontend/src/pages/CompatibilityHistoryPage.tsx`
- New compatibility history stylesheet
- Frontend static tests for archive layout and mobile safe area
- No backend API, database, or algorithm changes
