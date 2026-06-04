## 1. Journey Utilities

- [x] 1.1 Add a frontend pending-journey module for storing, reading, validating, expiring, clearing, and retrying guest bazi journeys in session storage.
- [x] 1.2 Add a safe post-auth target resolver that accepts only valid local user routes and rejects absolute URLs, protocol-relative URLs, malformed paths, and admin routes.
- [x] 1.3 Add unit tests for pending journey validation, expiration, clearing behavior, and unsafe redirect rejection.

## 2. Guest-To-Auth Continuation

- [x] 2.1 Update guest result-page login/register CTAs to persist the current bazi calculation input, chart context, display label, and intended action before navigation.
- [x] 2.2 Update login and registration pages to preserve safe `next` targets when switching between auth pages.
- [x] 2.3 After successful login/register, resolve pending bazi journeys by recalculating with the stored input under the authenticated user and navigating to the authenticated result target.
- [x] 2.4 Keep pending journey data available when authenticated recalculation fails, and show a recoverable error with a retry path.

## 3. Continuation And Archive Navigation

- [x] 3.1 Update profile continuation logic to choose the most recent item across bazi charts and compatibility readings by timestamp.
- [x] 3.2 Ensure the no-records profile state presents a start-new-analysis action instead of an empty continue link.
- [x] 3.3 Add or adjust the mobile bottom navigation archive destination so authenticated users can reach saved bazi and compatibility records without visiting profile first.
- [x] 3.4 Route anonymous mobile archive access through login/register with a safe return target.

## 4. Homepage And Profile Intent Clarity

- [x] 4.1 Rework homepage primary intent affordances so start bazi, continue saved analysis, and compatibility analysis are visually discoverable as active paths.
- [x] 4.2 Adjust longitude/true-solar-time copy so it accurately reflects the precision of the current input.
- [x] 4.3 Demote coming-soon profile features below active saved-analysis, history, compatibility, and settings actions.

## 5. Result Feedback Consistency

- [x] 5.1 Replace remaining native `alert()` usage in user-facing result/export/share flows with toast or inline retry feedback.
- [x] 5.2 Replace remaining native `confirm()` usage in user-facing saved-analysis delete flows with shared confirmation dialogs.
- [x] 5.3 Ensure AI generation, export, share, and delete controls show in-flight states and prevent duplicate submissions while requests are pending.
- [x] 5.4 Preserve partially generated content after streamed generation failures and expose a clear retry action for the failed request.

## 6. Verification

- [x] 6.1 Add or update frontend tests for guest calculation -> login/register -> authenticated result continuation.
- [x] 6.2 Add or update frontend tests for profile recent-continuation ordering across bazi and compatibility timestamps.
- [x] 6.3 Add or update frontend tests for mobile archive routing for authenticated and anonymous users.
- [x] 6.4 Add or update frontend tests confirming result export/share failures use shared feedback rather than native browser dialogs.
- [x] 6.5 Run frontend lint and build.
- [x] 6.6 Smoke test the core journey manually in a local browser: guest bazi calculation, login/register continuation, profile continue, mobile archive navigation, and result export failure handling.
