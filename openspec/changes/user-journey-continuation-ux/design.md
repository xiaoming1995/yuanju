## Context

The current bazi calculation endpoint accepts both anonymous and authenticated users. Anonymous calculations are saved with `user_id = NULL` for traffic visibility, but authenticated-only actions such as AI report generation require the chart row to belong to the logged-in user. Login and register currently navigate to `/` after success, so a guest who calculates a chart and then signs in can lose the immediate continuation path.

The frontend already has result pages, history pages, profile overview, bottom navigation, and shared UI primitives such as buttons, empty states, toasts, and confirm dialogs. This change should connect those pieces into a clearer journey rather than adding unrelated product features.

## Goals / Non-Goals

**Goals:**
- Preserve guest bazi calculation context through login/register and resume the intended action after authentication.
- Make post-auth routing deterministic and safe for local return targets.
- Provide a clearer "continue recent analysis" model across bazi charts and compatibility readings.
- Make saved readings easier to reach on mobile.
- Standardize result-page feedback for generation, export/share, retry, and confirmation states.
- Keep the implementation frontend-first unless backend ownership behavior is explicitly required.

**Non-Goals:**
- Changing bazi algorithms, compatibility scoring, yongshen reasoning, or AI prompt semantics.
- Adding billing, subscription, wallet, or usage-limit behavior.
- Reworking admin workflows beyond shared feedback primitives already used by the app.
- Adding a generic multi-step onboarding framework.
- Claiming anonymous database rows by chart ID without a verifiable claim token.

## Decisions

### Decision: Store pending guest journey client-side, then recreate under the authenticated user

When an anonymous user calculates a chart, the frontend will store a small pending journey object in session storage. The object will include a schema version, journey type, original `CalculateInput`, anonymous `chart_id` if present, display label, intended action, and creation timestamp.

After login/register, the app will resolve the pending journey. For bazi journeys, it will call `baziAPI.calculate` with the stored input while authenticated, producing a new user-owned chart. It will then navigate to the intended route with the authenticated chart ID and clear the pending object only after the authenticated chart is available.

Alternative considered: add a backend `claim` endpoint for anonymous chart rows. This would be cleaner for database deduplication but unsafe with the current API because anonymous chart IDs are not protected by a claim secret. A secure claim design would require issuing and storing a one-time claim token during calculation, which is more backend surface than this UX change needs.

### Decision: Post-auth return targets use safe local paths only

Login and register will support a `next` target from query string or route state. The target resolver will accept only same-origin relative paths starting with `/` and reject protocol URLs, protocol-relative URLs, admin paths for normal users, and malformed values. If no safe `next` exists, the resolver will use the pending journey; otherwise it will fall back to the profile or home page.

Alternative considered: always return to `/`. This is simple but preserves the current broken journey for guest-to-auth flows.

### Decision: Recent continuation is derived from available summaries, not fixed feature priority

Profile and mobile archive entry points will identify the most recent meaningful item by comparing bazi chart and compatibility reading timestamps. The UI can still expose separate "命盘记录" and "合盘记录" actions, but the primary continue action should not always prefer one feature over the other.

Alternative considered: keep bazi as the default continue target. That is predictable for MVP, but it ignores users who last used compatibility.

### Decision: Result actions share feedback semantics

Result-like pages will use consistent feedback categories:
- Toast for non-blocking success/failure messages.
- Inline error or retry panel for generation failures that affect the visible result.
- Confirm dialog for destructive actions.
- Loading/disabled states for in-flight generation, export, share, and delete actions.

Native `alert()` and `confirm()` will be removed from user-facing result workflows.

### Decision: Mobile archive access is a primary navigation concern

The bottom navigation will include a clear saved-reading/archive destination for authenticated users. For anonymous users, the same destination will route to login/register with a safe `next` target so the user lands in the archive after authentication.

Alternative considered: leave archive access under profile only. This keeps the nav smaller, but it hides one of the product's most repeated actions.

## Risks / Trade-offs

- Duplicate anonymous and authenticated chart rows may be created for the same guest calculation -> acceptable for this change because it avoids insecure claiming; future backend claim-token support can dedupe if needed.
- Session storage can be cleared or expire -> the user falls back to normal login/profile behavior and can recalculate from the homepage.
- Re-running calculation after login can fail -> keep the pending journey until retry or explicit dismissal, and show an inline recoverable error.
- Adding a fourth mobile nav item can reduce tap target space -> use short labels and existing icon-button sizing constraints, and verify mobile screenshots.
- More route-return logic can create open redirect risk -> centralize the safe local path resolver and test unsafe values.

## Migration Plan

1. Add frontend helpers for pending journey storage, expiration, and safe post-auth target resolution.
2. Update homepage/result auth CTAs to record pending bazi journeys and pass safe `next` targets.
3. Update login/register to resolve safe `next` and pending journeys after successful authentication.
4. Update profile and mobile navigation to expose a clearer recent/archive continuation path.
5. Replace remaining user-facing native alert/confirm result feedback with shared Toast, ConfirmDialog, inline errors, and retry states.
6. Add focused tests for route resolution, pending journey recovery, result action feedback, and mobile archive access.

Rollback is frontend-only for the default path: remove pending journey resolution and restore post-auth navigation to the previous default route. If a future backend claim-token endpoint is added, it should remain backward-compatible with the authenticated recalculation fallback.

## Open Questions

- Should pending guest journeys expire after the browser session only, or also by a short wall-clock timeout such as 24 hours?
- Should `/history` become the single archive shell for both bazi and compatibility, or should compatibility history stay as a separate route linked from the archive shell?
- Should the profile overview hide all coming-soon feature cards by default until at least one commercial feature is implemented?
