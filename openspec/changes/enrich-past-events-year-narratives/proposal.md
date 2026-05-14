## Why

Past-events year cards currently read like short reminders rather than meaningful yearly readings. Users can see tags and a brief sentence, but the default card body does not explain why that year matters, how the signal manifests in daily life, or what stance the user should take.

This change makes each rule-generated liunian card feel like a compact professional reading while keeping the module deterministic, fast, and free of per-year LLM token cost.

## What Changes

- Expand past-events year card narratives from short 2-3 sentence hints into medium-detail rule-generated readings.
- Compose each yearly narrative from structured sections: yearly tone, trigger source, real-life domain impact, ten-god force context, and practical stance.
- Keep hard event evidence such as clash, punishment, void, major formation, and use-god/jishen hits prominent in the reading.
- Preserve readability for non-professional users by using plain-language domain wording instead of dense technical terms.
- Keep the expandable evidence section as the place for detailed technical basis.
- Do not add per-year LLM calls or token usage.

## Capabilities

### New Capabilities
- `past-events-year-narrative-depth`: Rule-generated medium-detail yearly narratives for the past-events module.

### Modified Capabilities
- None.

## Impact

- Backend: `backend/pkg/bazi/event_narrative.go` will gain richer deterministic narrative composition.
- Backend tests: narrative tests will need broader coverage for length, distinctness, hard-evidence priority, and child/adult life-stage wording.
- Frontend: `frontend/src/pages/PastEventsPage.tsx` may need minor spacing or typography adjustments to keep longer text readable.
- API shape: no breaking response shape change is expected; existing `narrative` content becomes richer.
- Cost and performance: no new LLM calls, no database migration, and no intentional cache invalidation.
