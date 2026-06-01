# Past Events Mobile Stream Recovery Design

## Context

The past-events page can leave a dayun summary or year narrative in a permanent loading state on mobile. The observed flow is:

1. The user opens past-events calculation on a mobile device.
2. The initial years and sections render.
3. One dayun summary or section narrative is still streaming.
4. The user switches to another app.
5. After returning, that section remains stuck at "generating" even though the rest of the page is usable.

The current frontend relies on the stream callback reaching either `onDone` or `onError`. Mobile browsers may pause, kill, or detach a streaming request while the page is backgrounded without delivering either terminal callback. When that happens, `PastEventsPage` can keep `loading: true` summaries and an inflight guard that prevents natural recovery.

## Goal

Prevent any generated dayun summary or section narrative from staying in an endless loading state after mobile background/foreground transitions.

Successful behavior:

- Returning from the background checks for stale loading summaries.
- Stale generation state either recovers automatically or becomes a clear retry state.
- A stream request must always resolve to a terminal frontend state: success, error, aborted, or retryable interruption.
- Existing APIs and backend response formats remain compatible.

## Non-Goals

- Do not replace stream generation with a server-side job queue.
- Do not redesign the past-events page layout.
- Do not change bazi calculation, event narrative rules, or AI prompt behavior.
- Do not introduce a frontend UI framework.

## Approach

Use a two-layer frontend recovery strategy:

1. Page-level mobile recovery in `PastEventsPage`.
2. Stream-level terminal-state protection in the API client.

This keeps the change scoped to the failure mode while improving stream safety for both automatic page generation and manual per-section generation.

## Page-Level Recovery

`PastEventsPage` should track generation metadata for each dayun summary that can be loading:

- `startedAt`: when the current generation attempt began.
- `attempt`: number of frontend retry attempts for that section.
- `source`: whether the loading state came from initial page generation or manual section generation.

The page should listen for app return signals:

- `visibilitychange` when the document becomes visible.
- `pageshow`, including browser back-forward cache restore.
- `focus`, as a secondary desktop/mobile browser signal.
- `online`, so recovery can run after temporary connectivity loss.

On return, the page scans `summaries` for entries that are still `loading: true`.

Recovery rule:

- If a loading entry is younger than the stale threshold, leave it alone.
- If it is older than the stale threshold and has not exceeded the retry limit, clear the stale inflight marker and retry that dayun segment.
- If it has exceeded the retry limit, change it to a retryable interrupted state.

Recommended constants:

- Stale threshold: 10 seconds without a terminal stream state.
- Automatic retry limit: 1 retry per dayun segment per page lifetime.

The interrupted UI should not show endless loading. It should show concise copy such as "生成中断，点击重试" and reuse the existing manual generation action.

## Stream-Level Protection

`streamDayunSummaries` should accept optional cancellation and timeout controls while preserving the current callback contract.

Required behavior:

- Use `AbortController` internally.
- Maintain an inactivity timer that resets whenever a stream chunk is received.
- If the stream has no chunk for the timeout window, abort the request and call the error callback with a retryable timeout/interruption error.
- If the reader ends without a final event, still call the done callback only after all parsed data has been processed.
- Ensure `onDone` and `onError` are mutually terminal from the caller's perspective.

Recommended timeout:

- 45 seconds of stream inactivity before aborting.

This is intentionally longer than the page-level stale threshold because page recovery is about mobile lifecycle return, while stream timeout is about truly silent streams.

## Data Flow

Initial page flow:

1. Load years and base dayun metadata.
2. Mark non-future dayun summaries as loading with generation metadata.
3. Start the stream.
4. Each stream item clears loading for its segment.
5. Stream done clears inflight state.
6. Stream error clears inflight state and marks unresolved loading segments as retryable interruption.

Mobile return flow:

1. Browser fires visible/pageshow/focus/online.
2. Page scans loading summaries.
3. Stale summaries are retried once by dayun index.
4. If retry succeeds, the summary renders normally.
5. If retry fails or times out, the summary becomes retryable instead of loading forever.

Manual generation flow:

1. User clicks generate for one folded/future segment.
2. That segment enters loading with generation metadata.
3. Stream item clears loading.
4. Stream error, timeout, or mobile return recovery converts the section to retryable interruption.

## Error Handling

Use three distinct frontend states:

- `loading`: an active stream is expected to produce content soon.
- `interrupted`: a stale or aborted stream did not produce content, and the user can retry.
- `error`: a real backend or validation error occurred.

The implementation may store `interrupted` through the existing summary structure as a small status field or a specific error message, but the UI should not conflate it with permanent fatal failure.

The user-facing copy should be short:

- Interrupted: "生成中断，点击重试"
- Retry failure: "生成失败，请重试"

## Testing

Add focused frontend tests around `PastEventsPage` and the API stream helper.

Test cases:

- A summary that remains loading across a simulated foreground return is retried once.
- A stale loading summary becomes retryable after the retry limit is reached.
- A stream that never emits `done` or `error` triggers timeout handling.
- Manual per-section generation does not stay loading after stream interruption.

Where practical, use fake timers so tests do not wait for real timeout durations.

## Rollout

This can ship as a frontend-only bug fix. No migration is needed.

After implementation, verify manually on a mobile viewport:

1. Open the past-events page.
2. Start or wait for a dayun summary generation.
3. Background the browser or switch apps.
4. Return after at least 10 seconds.
5. Confirm the section either recovers or shows a retry action instead of endless loading.

## Self-Review

- No placeholder requirements remain.
- The scope is limited to mobile stream recovery for past-events generation.
- The design preserves existing API compatibility.
- The retry threshold and retry limit are explicit.
- The UI terminal states are distinct enough to avoid another permanent loading state.
