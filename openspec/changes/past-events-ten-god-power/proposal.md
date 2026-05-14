## Why

The past-events timeline already uses dayun and liunian ten-god names internally, but the user-facing result still reads mostly as event tags and generic yearly prose. This change makes each dayun and liunian explain the native's ten-god force in plain language, so users can understand not only "what may happen" but also "which force is driving it".

## What Changes

- Add structured ten-god power profiles for each dayun and liunian in the past-events years response.
- Classify ten gods into practical force groups: wealth, official/killing, seal, output, peer.
- Score and label the strength of each dayun and liunian force using heavenly stem, earthly branch main qi, dayun front/back phase, natal strength, and yongshen/jishen context.
- Expose a concise plain-language interpretation for each profile, avoiding raw terminology-heavy output.
- Update the past-events yearly narrative selector so ten-god force can guide the dominant theme without overriding stronger event evidence such as major clash, punishment, void, or dayun-liunian resonance.
- Update the dayun summary prompt input with ten-god force profiles so AI dayun summaries can describe the period's dominant force without increasing per-year LLM generation.
- Add focused frontend presentation for dayun and liunian ten-god power, using small labels and one-line explanations rather than dense technical panels.

## Capabilities

### New Capabilities

- `past-events-ten-god-power`: Defines structured dayun/liunian ten-god force profiles, scoring rules, API exposure, narrative usage, and frontend presentation for the past-events module.

### Modified Capabilities

- None.

## Impact

- Backend algorithm: `backend/pkg/bazi/event_signals.go`, related narrative helpers, and tests.
- Backend service/API: `backend/internal/service/report_service.go` past-events response structs and dayun summary prompt payload.
- Frontend: `frontend/src/pages/PastEventsPage.tsx` and API TypeScript types.
- No database schema changes are required.
- No new LLM calls are introduced; existing cached dayun summaries should remain usable, with regeneration only needed when users want summaries to include the new ten-god context.
