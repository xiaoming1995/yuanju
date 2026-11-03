## 1. Data and transaction

- [x] 1.1 Add a migration for `users.chart_query_count` with current-chart backfill.
- [x] 1.2 Increment count in the chart creation transaction and sum counts during miniapp account merge.

## 2. Admin display

- [x] 2.1 Return the cumulative count in admin user list and detail without changing existing `chart_count` semantics.
- [x] 2.2 Label and show cumulative queries and current charts clearly in user list and detail.

## 3. Verification

- [x] 3.1 Cover create, delete, guest, merge, backfill and admin response semantics with focused tests.
- [x] 3.2 Run relevant backend/frontend checks and validate OpenSpec change.
