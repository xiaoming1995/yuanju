## 1. Data Model And Migration

- [x] 1.1 Add additive storage for article quality score and structured quality reasons.
- [x] 1.2 Add additive storage for task-item quality score, quality reasons, and skip reason.
- [x] 1.3 Add article quality configuration storage for enabled flag, minimum score, source rules, bonus keywords, body policy, and optional AI flag.
- [x] 1.4 Add indexes needed for admin quality sorting/filtering.
- [x] 1.5 Add migration verification tests.

## 2. Quality Scoring Service

- [x] 2.1 Implement deterministic quality scoring from search rank, publish recency, body fetch status, title/snippet relevance, source rules, and content usability.
- [x] 2.2 Produce structured score reasons with point values and human-readable messages.
- [x] 2.3 Implement source blacklist hard-skip behavior and preferred-source boost behavior.
- [x] 2.4 Ensure external WeChat metrics are optional and never required for V1 collection.
- [x] 2.5 Add focused tests for score calculation, skip decisions, source rules, and missing metric handling.

## 3. Collection Pipeline Integration

- [x] 3.1 Run quality scoring after provider result parsing and before candidate insertion.
- [x] 3.2 Store score/reasons on inserted articles.
- [x] 3.3 Record low-score skipped candidates as task items with status `skipped`, score, and reason.
- [x] 3.4 Preserve URL deduplication and duplicate task item behavior.
- [x] 3.5 Add tests for enabled/disabled filtering and task count behavior.

## 4. Admin APIs

- [x] 4.1 Add get/update endpoints for article quality configuration.
- [x] 4.2 Extend admin article list/detail responses with quality score and reasons.
- [x] 4.3 Extend collection task item responses with quality score, reasons, and skip reason.
- [x] 4.4 Add admin list sorting/filtering by quality score where appropriate.
- [x] 4.5 Add handler tests for config updates and quality metadata visibility.

## 5. Admin Frontend

- [x] 5.1 Add quality filtering controls to the article collection/content configuration UI.
- [x] 5.2 Show quality score badges and reason summaries in task logs.
- [x] 5.3 Show skipped-by-quality task items with clear reasons.
- [x] 5.4 Show quality score/reasons in article detail collection metadata.
- [x] 5.5 Add article list sort/filter affordances for quality score.
- [x] 5.6 Add frontend tests for config wiring and quality display strings.

## 6. Verification

- [x] 6.1 Run backend repository/service/handler tests for quality scoring and collection.
- [x] 6.2 Run frontend lint/build/static tests.
- [x] 6.3 Manually verify collection with filtering disabled stores scores only.
- [x] 6.4 Manually verify collection with filtering enabled skips low-score items and logs reasons.
