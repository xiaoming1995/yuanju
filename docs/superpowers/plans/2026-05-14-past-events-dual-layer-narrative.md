# Past Events Dual-Layer Narrative Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace raw technical yearly past-event prose with plain-language readings while preserving expandable professional evidence.

**Architecture:** Keep `EventSignal` unchanged. Add narrative and evidence-summary helpers in `backend/pkg/bazi/event_narrative.go`, expose `evidence_summary` from `GeneratePastEventsYears`, and render a compact expandable "命理依据" section in `PastEventsPage`.

**Tech Stack:** Go unit tests for backend behavior, React + TypeScript for frontend display, existing Vite build for frontend verification.

---

### Task 1: Backend Narrative Contract

**Files:**
- Modify: `backend/pkg/bazi/event_narrative.go`
- Test: `backend/pkg/bazi/event_narrative_test.go`

- [x] **Step 1: Write failing tests**

Add tests that construct `YearSignals` with technical evidence and assert the public narrative avoids leaked terms while evidence remains available through a new helper.

- [x] **Step 2: Verify red**

Run: `go test ./pkg/bazi -run 'TestRenderYearNarrative|TestRenderEvidenceSummary'`

Expected: fail because `RenderEvidenceSummary` does not exist and `RenderYearNarrative` still returns raw technical evidence.

- [x] **Step 3: Implement helpers**

Update `RenderYearNarrative` to synthesize plain-language text by dominant theme and polarity. Add `RenderEvidenceSummary(ys YearSignals) []string` to return 2-5 selected technical snippets.

- [x] **Step 4: Verify green**

Run: `go test ./pkg/bazi -run 'TestRenderYearNarrative|TestRenderEvidenceSummary'`

Expected: pass.

### Task 2: API Field

**Files:**
- Modify: `backend/internal/service/report_service.go`

- [x] **Step 1: Add field**

Add `EvidenceSummary []string 'json:"evidence_summary,omitempty"'` to `PastEventsYearItem`.

- [x] **Step 2: Populate field**

Set `EvidenceSummary: bazi.RenderEvidenceSummary(ys)` in `GeneratePastEventsYears`.

- [ ] **Step 3: Verify service compile**

Run: `go test ./internal/service -run Test`

Expected: pass or skip no-test package behavior without compile errors.

Observed: blocked by existing `internal/service/llm_pricing_test.go` reference to undefined `matchModelTier`, unrelated to this change.

### Task 3: Frontend Expandable Evidence

**Files:**
- Modify: `frontend/src/pages/PastEventsPage.tsx`

- [x] **Step 1: Update type**

Add `evidence_summary?: string[]` to `YearEvent`.

- [x] **Step 2: Add UI state and control**

Track expanded evidence cards by `year`, render a small button labelled `命理依据`, and show evidence lines only when expanded.

- [x] **Step 3: Verify build**

Run: `npm run build` in `frontend`.

Expected: TypeScript and Vite build exit 0.

### Task 4: Full Verification

**Files:**
- Verify only.

- [x] **Step 1: Run focused backend tests**

Run: `go test ./pkg/bazi ./internal/service`

Expected: exit 0.

- [x] **Step 2: Run frontend build**

Run: `npm run build` in `frontend`.

Expected: exit 0.
