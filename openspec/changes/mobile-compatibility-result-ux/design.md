## Overview

This change keeps the compatibility calculation and report data unchanged, but changes how the result is presented. The page should answer "what does this mean for this relationship?" before showing professional chart details.

## UX Structure

Mobile order:

1. Conclusion hero: names, level, summary, tags, primary action.
2. Quick score overview: four score rows with compact bars and short labels.
3. Duration outlook: three relationship windows plus short summary.
4. Key risks and advice: generated report content when available, fallback to structural evidence when not.
5. Professional details: participants, four pillars, wuxing, and full evidences.
6. Full AI report: structured sections or raw report content.

Desktop keeps the same content, but benefits from the clearer section components and less inline styling.

## Component Approach

- Add small local helper components inside `CompatibilityResultPage.tsx`:
  - `ScoreOverview` for dimension rows.
  - `InsightPanel` for risks/advice/fallback guidance.
  - `ProfessionalDetails` wrapper using native `details`/`summary` for mobile-friendly progressive disclosure.
- Keep data derivation local to the page because no other page currently needs this exact presentation model.
- Keep all styling in `CompatibilityResultPage.css` and reduce inline styles where touched.

## Data Handling

- Use `reading.dimension_scores`, `reading.summary_tags`, `reading.duration_assessment`, and `detail.evidences` as the always-available baseline.
- Prefer `latest_report.content_structured` for richer summary, risks, advice, and dimensions.
- When no structured report exists, derive risk highlights from negative evidences and show a clear generate-report prompt.

## Mobile Layout Rules

- Result page must reserve bottom padding for the mobile bottom nav.
- Score rows must avoid tiny cards and horizontal overflow.
- Professional details should not dominate the first mobile viewport.
- Buttons and report actions must stay reachable above the bottom nav.

## Testing

- Add static tests that check the result page exposes the conclusion-first sections, score rows, progressive professional detail wrapper, and mobile bottom padding.
- Re-run lint, build, and the existing frontend node tests.
