## 1. Seasonal and 扶抑 assessment

- [x] 1.1 Add the `寅` early-spring thermal mapping with `初春偏寒` and required `火`, reusing the existing seasonal evidence states.
- [x] 1.2 Add centralized month-command relationship scoring and traceable `月令` evidence to the global 扶抑 assessment.
- [x] 1.3 Verify that day-master/month-command same, generating, draining, and controlling relationships produce deterministic directional evidence.

## 2. Shared-priority yongshen model

- [x] 2.1 Add the structured natal shared-priority yongshen alignment model and derive it from the independent thermal and 扶抑 outputs.
- [x] 2.2 Ensure alignment does not replace complete thermal, day-stem, or 扶抑 results and does not add a duplicate natal-grade score.
- [x] 2.3 Increment the natal assessment algorithm version so stale cached assessments are recalculated.

## 3. Product and AI surfaces

- [x] 3.1 Add shared-priority yongshen wording to vehicle-profile/natal-assessment payloads and frontend display.
- [x] 3.2 Add the alignment context to report and dayun prompt construction, with no output when no common element exists.

## 4. Verification

- [x] 4.1 Add backend regression tests for `丙子·庚寅·乙亥·丙戌`: early-spring fire requirement, month-command strength evidence, resolved 扶抑, and shared fire priority.
- [x] 4.2 Add negative and seasonal-control tests proving absent overlap is not aligned and existing non-`寅` rules retain their behavior.
- [x] 4.3 Update frontend/prompt tests and run targeted Go tests, frontend tests, and production build.
