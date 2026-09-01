## Why

The current Dayun phase area exposes `天干主事` and a Jin Bu Huan `吉/平/凶` result, but a non-specialist cannot immediately tell what road they will experience in each five-year period or what to focus on. The result is technically accurate but requires interpretation that the page should provide directly.

## What Changes

- Present the first and second five-year periods as direct, user-facing phase road conditions, while continuing to label the ten-year result as the composite road condition.
- Add a short actionable theme for every phase road condition, such as "快速推进", "稳步落地", "选择与节奏", "稳住调整", or "修整蓄力".
- Keep Jin Bu Huan and the governing Gan/Zhi available as secondary evidence instead of leading the default reading experience.
- Apply the same naming and theme treatment to the selected-Dayun timeline summary without adding phase details to every compact timeline card.

## Capabilities

### New Capabilities
- `dayun-phase-road-guidance`: Plain-language road and theme guidance for the front and back five-year Dayun phases.

### Modified Capabilities
- `dayun-timeline-design`: The selected-Dayun summary gains user-facing phase road and theme content while compact cards retain only the composite badge.

## Impact

- Affects `frontend/src/lib/dayunRoadPresentation.ts`, `frontend/src/pages/ResultPage.tsx`, `frontend/src/components/DayunTimeline.tsx`, and ResultPage styling and static tests.
- Reuses existing `qian_road` and `hou_road` labels and scores; no scoring, API contract, or backend data migration changes.
