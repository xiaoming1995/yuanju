## 1. Backend Data Contract

- [x] 1.1 Add tests for day-master relation matrix generation, including day stem labeled as "日主 / 日元".
- [x] 1.2 Add tests for hidden-stem pairing so each hidden stem maps to its own ten-god label without index mismatch.
- [x] 1.3 Define structured relation matrix types in the bazi package or result model.
- [x] 1.4 Add deterministic ten-god relation descriptors and grouped plain-language summaries.
- [x] 1.5 Populate the relation matrix during `Calculate`.
- [x] 1.6 Ensure old chart snapshots can derive or lazily recalculate relation data without breaking existing fields.

## 2. Frontend Result Page

- [x] 2.1 Add static tests for the "命主十神关系" module, including day-master reference text and heavenly-stem rows.
- [x] 2.2 Add static tests for hidden-stem rows and mobile no-horizontal-overflow layout expectations.
- [x] 2.3 Extend the `BaziResult` TypeScript shape with optional relation matrix data.
- [x] 2.4 Add frontend fallback derivation from existing raw fields when structured relation data is absent.
- [x] 2.5 Render the day-master summary and heavenly-stem relation cards near the basic chart section.
- [x] 2.6 Render hidden-stem relation cards with concise ten-god explanations.
- [x] 2.7 Style desktop and mobile layouts without increasing density in the existing professional grid.
- [x] 2.8 Keep "基本排盘" as the first detail section and render ten-god explanation after it.

## 3. Verification

- [x] 3.1 Run backend bazi tests focused on ten-god relation data.
- [x] 3.2 Run frontend static tests for result-page relation rendering.
- [x] 3.3 Run `npm run lint`, frontend static tests, backend package tests, and `npm run build`.
- [x] 3.4 Verify `/` to result-page flow in a mobile viewport and confirm no horizontal overflow.
