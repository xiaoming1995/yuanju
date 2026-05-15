## Why

The current result page only approximates the provided "大运时间轴" mockup through incremental style tweaks, so desktop and mobile layouts still differ from the approved design in structure, density, metadata placement, and card proportions. A dedicated change is needed to treat the mockup as the source of truth and rebuild the timeline module around pixel-aligned desktop and mobile presentation.

## What Changes

- Rebuild the Bazi result "大运时间轴" section to match the supplied desktop and mobile mockup layout.
- Render desktop as a single large timeline panel with header metadata, ten compact Dayun cards, a selected/current indicator, a nested Liunian card panel, Dayun summary, and footer disclaimer.
- Render mobile as a dedicated phone-style compact layout with top title bar, metadata block, two-column Dayun cards, two-column Liunian cards, and a bottom Dayun summary control.
- Align colors, borders, typography scale, card spacing, badges, corner ribbons, active state, and rounded containers with the reference image.
- Ensure current year, transition year, and focus year indicators use the visual language shown in the mockup.
- Preserve existing interactions such as selecting Dayun cards, opening Liuyue details, and Shensha annotation access where applicable.
- Normalize the Dayun/Liunian display contract so the UI can show ten timeline cards when the mockup requires ten cards.

## Capabilities

### New Capabilities

- `dayun-timeline-design`: Defines the visual and behavioral contract for the redesigned Dayun/Liunian timeline on desktop and mobile result pages.

### Modified Capabilities

- None.

## Impact

- Frontend Dayun timeline component: `frontend/src/components/DayunTimeline.tsx`
- Frontend result page structure: `frontend/src/pages/ResultPage.tsx`
- Frontend result page styles: `frontend/src/pages/ResultPage.css`
- Frontend static and browser verification tests for desktop/mobile timeline fidelity
- Possible backend Bazi Dayun sequence adjustment if current calculation cannot provide ten displayable Dayun periods
- No new third-party UI libraries; continue using React, TypeScript, and CSS variables
