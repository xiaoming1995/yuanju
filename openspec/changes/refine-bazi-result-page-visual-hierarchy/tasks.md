## 1. Result Page Information Architecture

- [x] 1.1 Reorder page-level result sections into core conclusion, current-state summary, and progressive-reading groups without changing result data or routes.
- [x] 1.2 Move the historical-figure reference block after the summary and trend reading path, retaining its simple and professional content behavior.
- [x] 1.3 Preserve a stable summary layout when vehicle or current-road data is unavailable, without rendering empty panels.

## 2. Visual System Consolidation

- [x] 2.1 Refine the verdict, vehicle, and road summaries around one shared heading, spacing, tag, divider, and action treatment.
- [x] 2.2 Reduce stacked-card and duplicate decorative treatments in the overview while preserving accessible modal and navigation controls.
- [x] 2.3 Consolidate result-overview CSS rules and add responsive single-column behavior for narrow viewports.

## 3. Regression Coverage and Visual QA

- [x] 3.1 Update focused result-page tests for section order, missing-data fallback, professional-mode continuity, and concise summary classes.
- [x] 3.2 Run frontend tests and production build, then inspect desktop and mobile result pages in both reading modes for hierarchy, alignment, overflow, and modal regressions.
