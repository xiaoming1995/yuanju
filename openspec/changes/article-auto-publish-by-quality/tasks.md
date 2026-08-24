## 1. Data Model And Migration

- [x] 1.1 Add auto-publish configuration fields to article collection config storage.
- [x] 1.2 Normalize and validate auto-publish threshold and per-run cap.
- [x] 1.3 Add structured task item visibility for auto-published outcomes if needed.
- [x] 1.4 Add migration verification tests.

## 2. Backend Auto-Publish Flow

- [x] 2.1 Extend collection config model/repository get/update paths with auto-publish fields.
- [x] 2.2 Implement auto-publish eligibility checks after successful new candidate insertion.
- [x] 2.3 Publish only newly inserted candidates that meet threshold, body, and per-run cap rules.
- [x] 2.4 Write automatic publish audit events with system semantics.
- [x] 2.5 Preserve duplicate, skipped, failed, and manually reviewed candidate behavior.
- [x] 2.6 Add service tests for disabled, threshold pass/fail, missing body, per-run cap, and duplicate cases.

## 3. Admin APIs

- [x] 3.1 Extend collection config get/update endpoints with auto-publish fields.
- [x] 3.2 Extend collection task item responses with auto-publish visibility if structured fields are added.
- [x] 3.3 Add handler tests for config persistence and task item visibility.

## 4. Admin Frontend

- [x] 4.1 Add auto-publish controls under the article collection scheduling UI.
- [x] 4.2 Clearly separate auto-publish threshold from quality filter threshold.
- [x] 4.3 Show auto-publish summary state after saving configuration.
- [x] 4.4 Show auto-published outcomes in task logs.
- [x] 4.5 Add frontend tests for auto-publish config wiring and display strings.

## 5. Verification

- [x] 5.1 Run backend repository/service/handler tests for article auto-publish.
- [x] 5.2 Run frontend lint/build/static tests.
- [ ] 5.3 Manually verify disabled auto-publish leaves high-score articles as candidates.
- [ ] 5.4 Manually verify enabled auto-publish publishes only high-score articles with body content.
- [ ] 5.5 Manually verify per-run auto-publish cap is respected.
