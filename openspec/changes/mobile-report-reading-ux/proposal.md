## Why

AI reports currently render as long-form content with a mode switch, but mobile users need a faster way to understand the key judgment before reading every section. After improving input, result, and archive flows, report readability is the next high-impact mobile experience gap.

## What Changes

- Add a summary-first report reading layer for structured Bazi reports.
- Render report chapters as mobile-friendly expandable sections.
- Add a compact terminology strip for common Bazi concepts.
- Add a report action bar linking to history, past-events analysis, PDF export, and re-calculation where applicable.
- Preserve existing AI report data contracts and generation behavior.

## Capabilities

### New Capabilities
- `mobile-report-reading-ux`: Mobile-first reading hierarchy for Bazi AI reports.

### Modified Capabilities
- None.

## Impact

- Frontend result page: `frontend/src/pages/ResultPage.tsx`
- Frontend result page styles: `frontend/src/pages/ResultPage.css`
- Frontend static tests for report reading hierarchy and mobile safe area
- No backend API, prompt, database, or report generation changes
