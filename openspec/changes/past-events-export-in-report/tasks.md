## 1. Read-Only Export Data

- [x] 1.1 Add backend service logic to aggregate existing cached dayun summaries for a chart without invoking any LLM generation path.
- [x] 1.2 Combine cached dayun summary rows with deterministic year/dayun metadata needed for export ordering, age ranges, year ranges, and labels.
- [x] 1.3 Add an authenticated read-only API endpoint for past-events export data and enforce chart ownership checks.
- [x] 1.4 Add backend tests proving the export data endpoint returns cached segments only and does not generate missing segments.

## 2. Frontend Export Models

- [x] 2.1 Add frontend API types and client method for reading past-events export data.
- [x] 2.2 Add a shared view-model/helper that filters exportable generated segments and formats year narratives, themes, labels, and optional evidence.
- [x] 2.3 Add unit tests for export filtering: generated, missing, loading, interrupted, folded, and partial-cache states.

## 3. Past Events Page Export

- [x] 3.1 Add a past-events page export action with copy that indicates it exports already generated content.
- [x] 3.2 Add an option for including or omitting命理依据/年份信号 in the past-events PDF.
- [x] 3.3 Build a print/PDF-specific past-events layout that excludes navigation, retry, expand/collapse, loading, and other interactive controls.
- [x] 3.4 Wire desktop export through the existing browser print flow and mobile export through the existing PDF fallback pattern.

## 4. Main Report Export Integration

- [x] 4.1 Load read-only past-events export data on the result page before exporting the main report.
- [x] 4.2 Append a "过往年运回看" section to the main report print layout only when exportable past-events data exists.
- [x] 4.3 Keep the main report appended section concise by omitting detailed evidence summaries by default.
- [x] 4.4 Ensure no empty past-events section or generation prompt appears when there is no cached past-events content.
- [x] 4.5 Load the same read-only past-events export data before saving the main report share image.
- [x] 4.6 Append a concise "过往年运回看" module to the share image only when exportable past-events data exists.
- [x] 4.7 Keep the share image module concise by omitting detailed evidence summaries and yearly signal lists.

## 5. Verification

- [x] 5.1 Verify the past-events page PDF includes only generated content and handles partial generation cleanly.
- [x] 5.2 Verify the main report PDF appends generated past-events content when cache exists.
- [x] 5.3 Verify the main report PDF remains unchanged when cache does not exist.
- [x] 5.4 Run frontend unit tests and build.
- [x] 5.5 Run relevant backend tests for the new read-only export endpoint.
- [x] 5.6 Verify the share image renders generated past-events content only when cache exists.
