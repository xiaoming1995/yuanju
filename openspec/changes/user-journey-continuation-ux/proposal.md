## Why

Yuanju's MVP already supports bazi calculation, AI reports, history, compatibility, and profile workflows, but the user journey is still fragmented around the highest-value moments: guest calculation, login/register, continuing an analysis, returning to prior readings, and recovering from generation/export failures. This change improves conversion and retention by making the user's current analysis follow them across authentication, result pages, history, profile, and mobile navigation.

## What Changes

- Preserve a guest user's in-progress bazi chart across login/register and return them to the intended next action after authentication.
- Add a consistent continuation model for result pages, history, and profile so users can resume the most relevant unfinished or recent analysis.
- Improve mobile return paths by making saved readings and analysis archives easier to reach from primary navigation.
- Standardize result-page action feedback for report generation, export/share failures, destructive confirmations, and retry states.
- Clarify homepage and profile entry points around user intent: start a new chart, continue a previous analysis, view records, or perform compatibility analysis.
- Do not change bazi calculation algorithms, compatibility scoring, AI prompt semantics, provider configuration, billing behavior, or database ownership rules except where persistence is necessary for user journey continuity.

## Capabilities

### New Capabilities
- `user-journey-continuation`: Covers guest-to-auth continuation, post-auth return targets, recent-analysis resume behavior, mobile archive access, and unified result action feedback.

### Modified Capabilities

## Impact

- Frontend pages: home, login, register, profile, bazi result, history, past-events, compatibility entry/result/history, and mobile navigation.
- Frontend shared UI: route guards, auth redirect handling, local pending analysis storage, result action bars, toast/confirm/error feedback components.
- Backend/API: may require lightweight ownership or claim endpoints if existing chart/report creation APIs cannot safely attach a pre-auth analysis to the authenticated user.
- Tests: frontend unit/integration coverage for post-auth return, pending chart continuation, mobile archive navigation, and result action feedback states; backend tests only if claim/ownership API behavior changes.
