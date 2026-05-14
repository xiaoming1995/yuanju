## 1. Backend Profile API

- [x] 1.1 Add user center response models for account info, stats, recent chart summaries, recent compatibility summaries, and feature entry states.
- [x] 1.2 Add repository queries for chart count, AI report count, compatibility reading count, recent charts, and recent compatibility readings scoped to the authenticated user.
- [x] 1.3 Add service logic to assemble the profile overview and mark wallet/PDF-template entries as unavailable placeholders.
- [x] 1.4 Add an authenticated handler and route such as `GET /api/user/profile`.
- [x] 1.5 Add backend tests for authenticated ownership scoping and unauthenticated rejection.

## 2. Frontend Profile Page

- [x] 2.1 Add TypeScript API types and a `userAPI.profile()` request method.
- [x] 2.2 Create `ProfilePage` with account summary, usage statistics, recent records, and quick actions.
- [x] 2.3 Display recharge/credit and PDF template customization cards as clearly unavailable future features.
- [x] 2.4 Handle loading, empty, and error states without breaking the rest of the app.

## 3. Navigation Integration

- [x] 3.1 Register `/profile` in `App.tsx` behind the ordinary user layout.
- [x] 3.2 Update `Navbar` so the logged-in nickname or a dedicated link enters the profile page.
- [x] 3.3 Update `BottomNav` so logged-in users see “我的” pointing to `/profile`, while history remains accessible from the profile page.
- [x] 3.4 Ensure unauthenticated access to `/profile` leads to login instead of private data rendering.

## 4. Verification

- [x] 4.1 Run backend tests covering the new profile aggregation behavior.
- [x] 4.2 Run frontend build or relevant frontend tests.
- [x] 4.3 Manually verify a logged-in profile page shows account stats and links to existing history/detail pages.
- [x] 4.4 Manually verify recharge and PDF template cards do not trigger payment or template changes.
