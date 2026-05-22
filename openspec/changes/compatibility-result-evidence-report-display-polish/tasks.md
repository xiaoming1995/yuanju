## 1. Evidence Disclosure

- [x] 1.1 Add static tests for default-open core claim evidence, collapsed claim previews, and disclosure state copy.
- [x] 1.2 Update `EvidenceLinkedClaims` so the first claim shows reasoning and linked evidence by default.
- [x] 1.3 Add concise preview text for collapsed claim evidence items.
- [x] 1.4 Update evidence disclosure labels so expanded and collapsed states are clear.
- [x] 1.5 Refine evidence CSS for compact default-open cards and mobile scanability.

## 2. Deep Report Display

- [x] 2.1 Add static tests for absent, loading, error, raw, and structured deep report states.
- [x] 2.2 Move or consolidate the report generation action so exactly one visible primary generation action exists when no report is present.
- [x] 2.3 Replace the large empty deep report card with a compact actionable state when no report exists.
- [x] 2.4 Add loading and error presentation inside the deep report module.
- [x] 2.5 Refine structured and raw report styling for readable section hierarchy.

## 3. Professional Details Support

- [x] 3.1 Add static tests for collapsed professional summary visibility and expanded detail availability.
- [x] 3.2 Update professional details summary copy to communicate available data before expansion.
- [x] 3.3 Refine professional details CSS so it reads as supporting evidence rather than another empty block.

## 4. Verification

- [x] 4.1 Run focused compatibility result frontend tests.
- [x] 4.2 Run `npm run lint` from `frontend/`.
- [x] 4.3 Run `npm run build` from `frontend/`.
- [x] 4.4 Smoke test `/compatibility/:id` in the in-app browser when local auth/data are available.
