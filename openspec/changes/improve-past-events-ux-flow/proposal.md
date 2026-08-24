## Why

The past-events page already separates deterministic year signals from AI dayun summaries, but the UI does not make that distinction clear enough. Users can see loading, folded, cached, interrupted, and future states in the same timeline without a clear sense of what is ready, what is generating, and what action is expected next.

This change improves the existing progressive-generation experience without changing the bazi algorithms, the default AI generation scope, or the token-saving two-click pattern for future dayun segments.

## What Changes

- Add a clearer page-level status summary for the past-events flow, including deterministic year readiness, AI generation progress, interrupted segments, and retry availability.
- Make the current dayun and current year easier to find and read first.
- Clarify dayun segment states with user-facing labels such as ready, generating, cached/generated, future not generated, and interrupted.
- Preserve the future-dayun two-step behavior while making the actions explicit: first reveal algorithmic year signals, then optionally generate AI commentary for that segment.
- Reduce default cognitive load in year cards by prioritizing readable conclusions and keeping technical evidence available as secondary detail.
- Ensure newly supported event signals such as `夹拱` are visible in past-events signal chips instead of being silently dropped.
- Keep backend API contracts, AI prompt strategy, cache semantics, and default auto-generation scope unchanged unless implementation reveals a small compatibility fix is required.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `past-events-progressive-generation`: Clarify user-facing progressive-generation states, current-period focus, future-segment affordances, readable year-card presentation, and signal chip visibility.

## Impact

- Frontend past-events page and related UI components under `frontend/src/pages` and any extracted components/styles.
- Frontend API handling only if needed to expose already-available stream/cache/error states more clearly.
- Past-events signal label mapping, including `夹拱`.
- Tests covering current-period focus, segment state labels, two-step future generation behavior, interrupted/retry state, and signal chip rendering.
- No database migration, backend route change, AI provider change, or default token-cost increase is expected.
