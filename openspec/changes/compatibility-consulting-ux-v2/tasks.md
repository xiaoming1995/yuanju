## 1. Backend Data Contract

- [ ] 1.1 Add compatibility context fields to backend model types, create request types, detail responses, and history item types.
- [ ] 1.2 Add a database migration for nullable/defaulted `relationship_stage` and `primary_question` fields on compatibility readings.
- [ ] 1.3 Update compatibility repository create/read/history queries to persist and return normalized context fields.
- [ ] 1.4 Add service-level normalization for unknown or missing context values to `general`.
- [ ] 1.5 Add backend tests covering create-with-context, create-without-context, unknown-value fallback, and history context output.

## 2. AI Prompt and Structured Report

- [ ] 2.1 Extend compatibility prompt data with relationship stage and primary question labels.
- [ ] 2.2 Update the compatibility prompt fallback to branch report emphasis by primary question.
- [ ] 2.3 Extend structured report models if needed for question-specific sections without breaking existing reports.
- [ ] 2.4 Add service tests proving context is included in prompt data and legacy readings still generate reports.

## 3. Frontend Input Flow

- [ ] 3.1 Extend frontend API types and create-reading request payload with optional relationship context.
- [ ] 3.2 Add relationship stage and primary question controls to the compatibility input page.
- [ ] 3.3 Keep context selection lightweight on mobile and preserve existing two-profile input flow.
- [ ] 3.4 Add static/frontend tests asserting context controls and payload wiring.

## 4. Frontend Decision Result UX

- [ ] 4.1 Update the result page top hierarchy to show context, decision headline, core contradiction, and next actions before score details.
- [ ] 4.2 Rephrase score labels into user-question language while keeping current score values and keys.
- [ ] 4.3 Render duration windows as stage tasks with risk, trigger, and action language.
- [ ] 4.4 Keep professional evidence and claim-linked rationale expandable behind secondary controls.
- [ ] 4.5 Ensure only one report-generation primary action is visible when no report exists.
- [ ] 4.6 Add result page UX tests for decision-first order, question-style score labels, stage task language, and evidence expansion markers.

## 5. Verification

- [ ] 5.1 Run focused backend compatibility tests for model, repository, service, handler, and bazi packages.
- [ ] 5.2 Run database migration tests with Docker access.
- [ ] 5.3 Run frontend static UX tests and mobile page QA tests.
- [ ] 5.4 Run frontend build when dependencies are available.
- [ ] 5.5 Manually inspect one new reading and one legacy reading to confirm fallback behavior and page hierarchy.
