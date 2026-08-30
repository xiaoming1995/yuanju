## 1. Historical Reference Data and Access

- [x] 1.1 Add a migration and backend model for curated Ming Ge historical references, including source metadata, birth-data precision, turning-point data, optional Dayun data, display order, and review status.
- [x] 1.2 Implement repository validation that permits publication and Dayun display only when the required source and verification fields are complete.
- [x] 1.3 Add an unauthenticated read API that returns at most two published references for an exact Ming Ge and omits unavailable Dayun data.
- [x] 1.4 Add backend tests for publication filtering, Ming Ge filtering, maximum result count, and Dayun display gates.
- [x] 1.5 Backfill missing `ming_ge` and `ming_ge_desc` in valid historical result snapshots without recalculating unrelated Bazi fields, with regression coverage for old snapshots.

## 2. Editorial Administration

- [x] 2.1 Add authenticated admin CRUD routes and request validation for historical references without modifying or exposing legacy generic celebrity records.
- [x] 2.2 Add an admin management view for source details, review status, birth-data confidence, turning points, and Dayun-display eligibility.
- [x] 2.3 Surface actionable validation feedback when an administrator attempts to publish incomplete Dayun data.

## 3. Result-Page Experience

- [x] 3.1 Add a frontend API client and result-page state for loading historical references only when `ming_ge` is present.
- [x] 3.2 Render an unframed “古人映照” section in the Ming Ge area with concise identity and historical-memory content in ordinary mode.
- [x] 3.3 Render source, turning-point, and API-authorized Dayun correspondence only in professional mode, with a fixed statement that the content is not a user match, template, or prediction.
- [x] 3.4 Preserve result rendering when a chart has no Ming Ge, a category has no published references, or the read request fails.

## 4. Curated Initial Content

- [x] 4.1 Define the editorial review checklist and acceptable source standard for historical identity, achievement, birth-data precision, and Dayun assertions.
- [x] 4.2 Prepare and review one or two source-backed draft references for each qualified common Ming Ge category; leave unsupported categories unpublished.
- [x] 4.3 Publish only entries that meet the review and Dayun-display requirements, and verify their public payloads contain no draft or unverifiable fields.

## 5. Verification

- [x] 5.1 Add focused backend, handler, and frontend tests for public filtering, validation failures, ordinary/professional rendering, and request-failure fallback.
- [x] 5.2 Run the relevant Go tests, frontend static tests, and production build.
- [x] 5.3 Run strict OpenSpec validation and manually inspect desktop and mobile result-page layouts for concise ordinary-mode content, complete professional evidence, and text fit.
