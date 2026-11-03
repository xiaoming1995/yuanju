## 1. Scoped chart query

- [x] 1.1 Add validated optional `user_id` to admin chart list and apply it to count and rows.
- [x] 1.2 Test user ID filtering, unfiltered behavior, and invalid IDs.

## 2. Admin navigation

- [x] 2.1 Link user names and detail to the scoped chart page.
- [x] 2.2 Read the URL scope on the chart page; preserve date filtering, pagination, empty state, and full-site exit.

## 3. Verification

- [x] 3.1 Build and lint touched frontend, run focused Go tests, and validate the change.

## 4. Deleted chart history

- [x] 4.1 Add durable minimal query history with safe backfill and cumulative count alignment.
- [x] 4.2 Write history in chart creation, derive deleted state from chart absence, and transfer it on account merge.
- [x] 4.3 Return paginated scoped query history including deleted status; keep unscoped charts compatible.
- [x] 4.4 Render deleted history without chart content or a detail link.
- [x] 4.5 Test migration, creation, deletion, merge, API and frontend; validate OpenSpec.

## 5. Count entry

- [x] 5.1 Link each cumulative chart query count to the same user-scoped chart history page, with clear keyboard focus and accessible text.
- [x] 5.2 Verify build, OpenSpec validation, and the local list-to-history interaction.
