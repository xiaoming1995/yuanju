## Why

The current Dayun view places the ten-year composite road label beside the front-five and back-five Jin Bu Huan labels without explaining that they use different scopes. A user can therefore see "泥路" together with two "施工路段" labels and reasonably assume the calculation contradicts itself.

## What Changes

- Present the active Dayun as a readable three-layer summary: ten-year composite road condition, plain-language guidance, and front/back phase prompts.
- Rename front-five and back-five labels as Jin Bu Huan phase prompts instead of presenting them as equivalent road grades.
- Replace the ambiguous aggregate sentence so it never calls the phase labels the "overall road condition".
- Reduce the Dayun timeline card and summary-strip density while preserving existing period selection, Liunian drill-down, Shen Sha annotations, and professional evidence access.
- Add a compact explanation entry point that distinguishes the ten-year composite score from the front/back phase rating.

## Capabilities

### New Capabilities
- `dayun-road-meaning`: Plain-language display rules that distinguish a composite ten-year road result from Jin Bu Huan front/back phase prompts.

### Modified Capabilities
- `dayun-timeline-design`: The Dayun summary strip and selected-period display need a clearer hierarchy and less dense, more understandable road-related content.

## Impact

- Frontend: `frontend/src/pages/ResultPage.tsx`, `frontend/src/components/DayunTimeline.tsx`, and their result-page styles and static tests.
- Backend: Dayun roadmap summary text will be clarified while preserving additive API fields and existing scoring.
- No database migration, route, or scoring-weight change is required.
