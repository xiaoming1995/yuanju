# Screenshot QA Checklist · UX System Result Page Polish

## Required Viewports

| Viewport | Purpose |
|---|---|
| 390px width | Mobile result-page readability, no horizontal overflow, reachable primary CTA |
| 1440px width | Desktop hierarchy, non-sticky desktop action layout, content rhythm |

## Result Page · 390px

- [ ] First screen shows chart identity and four-pillar overview.
- [ ] First screen shows a plain-language core conclusion.
- [ ] Primary AI interpretation action is reachable without scrolling to the bottom.
- [ ] Bottom action area does not cover BottomNav or page content.
- [ ] Segmented navigation scrolls horizontally if needed without causing page-level horizontal overflow.
- [ ] Professional modules keep readable text sizes.

## Result Page · 1440px

- [ ] Hero content has clear hierarchy and does not look like a marketing landing page.
- [ ] Primary and secondary actions are visible near the hero.
- [ ] Segmented navigation anchors to stable detail sections.
- [ ] Desktop does not use a sticky bottom CTA.
- [ ] Existing report generation, export, and past-events paths remain visible.

## Regression Checks

- [ ] Existing chart detail modules remain visible below the decision-first hero.
- [ ] AI report missing state still offers a generation action.
- [ ] Saved history route still loads a chart result.
- [ ] Export controls remain available when a report exists.
- [ ] Past-events entry remains disabled for guests and usable for saved charts.
