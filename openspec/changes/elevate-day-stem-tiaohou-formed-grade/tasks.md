## 1. Day-Stem Formation Model

- [x] 1.1 Extend the versioned day-stem adjustment assessment with formation, foundation tier, and auditable foundation score fields; bump the natal assessment version.
- [x] 1.2 Implement exact required-stem coverage and heaven-and-earth evidence checks for formed, partial, unformed, and unavailable outcomes.
- [x] 1.3 Add a dedicated `日干调候成格` evidence record and incorporate its explicit foundation contribution into natal scoring.

## 2. Pattern and Vehicle Assessment

- [x] 2.1 Add a separate high-foundation summary to the natal pattern assessment without changing the generic primary-pattern root/制化 quality semantics.
- [x] 2.2 Preserve existing 制化配合 and critical-break rules, and update vehicle-grade/summary composition to show high foundation alongside retained thermal,扶抑, and break constraints.
- [x] 2.3 Update vehicle tags and evidence labels so `主格结构` and `日干调候成格·高格基础` cannot be conflated in API output.

## 3. User and AI Presentation

- [x] 3.1 Update the result-page assessment components to lead with day-stem formed/high-foundation status and separately render primary-pattern structure and 制化配合.
- [x] 3.2 Update AI report and flow-year prompt context to pass the formation conclusion, its visible/hidden evidence, main-pattern quality, and 制化 evidence as distinct inputs.

## 4. Verification

- [x] 4.1 Add backend tests for the 1995-10-12 午时 chart asserting `甲`透、`壬`藏 produces formed/high foundation and a dedicated evidence entry.
- [x] 4.2 Add backend table tests for only-visible, only-hidden, incomplete-required, and unrelated-extra-stem cases.
- [x] 4.3 Update vehicle/profile and report-service tests to assert the separated foundation and main-pattern semantics.
- [x] 4.4 Add or update frontend rendering tests, run Go tests and the frontend build, and verify the target chart through the calculate/history API response.
