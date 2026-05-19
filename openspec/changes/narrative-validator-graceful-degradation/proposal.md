## Why

Phase 1 of the yongshen-priority-realignment program shipped 喜忌十神 prompt injection and brought the past-events empty-year rate from 60% → 3.3% (18x reduction). The remaining gap is concentrated on a single failure mode: AI writes a high-frequency 命理 term (most commonly 伏吟) that the year's evidence does not attest to, and `ValidateYearNarrative` clears the **entire** narrative.

Two structural problems make these the last hard-to-close 3%:

1. **All-or-nothing validator**: When AI writes 150 characters of which 145 are well-grounded and one phrase mentions an unattested 伏吟, the validator discards all 150 characters. The user sees an empty card or a chip-only fallback, even though the bulk of the narrative was correct.

2. **神煞 vs 关系词 treated identically**: The 24 神煞 names and the 5 hard relation words (伏吟/反吟/三合/三会/大运合化) are scored by the same substring check. A drifted 神煞 reference ("驿马动象") is treated the same as a falsely-claimed 反吟. Different stakes warrant different responses.

The lesson from Phase 1's reverted Patch 4 (per-year `allowed_keywords` prompt-side whitelist) is that **prompt-side restriction backfires** — the AI reads enumerated lists as suggestions, not restrictions. The right place to recover from AI overconfidence is the **validator** side, where the action is deterministic.

## What Changes

- **Sentence-level granularity**: `ValidateYearNarrative` returns a richer result that identifies which sentence(s) contain the offending term. The caller (`GenerateDayunSummariesStream`) clears only those sentences, preserves the rest, and re-joins. Narratives with one offending sentence keep 80-90% of their content visible to the user.

- **Bifurcate the keyword classes by stakes**:
  - **Hard set** (5 terms): 伏吟, 反吟, 三合, 三会, 大运合化. These are exact干支 relationships; AI fabricating them produces materially wrong 命理 claims. Keep strict substring validation; offending sentence is cleared.
  - **Position set** (3 terms): 用神位, 忌神位, 喜神位. Used as structural anchors; misuse misleads users about chart structure. Keep strict; offending sentence is cleared.
  - **Event marker set** (4 terms): 受冲, 受刑, 双重命中, 力度倍增. Algorithm-internal vocabulary; AI usage outside evidence is a misrepresentation. Keep strict; offending sentence is cleared.
  - **神煞 set** (24 terms): 驿马, 桃花, 华盖, 白虎, 丧门, 吊客, 灾煞, 流霞, 天医, 天喜, 天乙, 天德, 月德, 文昌, 太极, 福星, 红艳, 孤辰, 寡宿, 羊刃, 亡神, 劫煞, 披麻, 咸池, 勾绞, 国印. Per 《滴天髓》"神煞末也" doctrine, these are decorative rather than load-bearing. Soft-warn: sentence is kept with an inline suffix marker `(注：未在本年算法 evidence 中识别)` so users see the content but know it's not algorithmically grounded.

- **Aggregate diagnostics**: When any sentence is cleared or soft-warned, log a structured event (call_type, dayun, year, term, action) so the cumulative impact on narratives is observable. Replaces the binary "整段丢弃" log.

- **No prompt changes** — Phase 1's prompt is the equilibrium; we only change post-AI handling.

- **No cache invalidation** — only newly generated rows benefit. Old rows remain at v2-yongshen-shishen baseline (current state).

## Capabilities

### Modified Capabilities
- `yongshen-driven-flow-year-judgment` (from Phase 1): the `ValidateYearNarrative` contract is extended from `(bool, string)` to a structured result with cleared-sentence indices and soft-warn annotations.

### New Capabilities
- `narrative-graceful-degradation`: Sentence-level narrative preservation with bifurcated keyword strictness.

## Impact

- Backend: `internal/service/year_narrative_validate.go` gains a richer return type. `report_service.go::GenerateDayunSummariesStream` adopts the new sentence-aware caller pattern.
- Backend tests: ~10 new `ValidateYearNarrative` cases (sentence split correctness, hard-set strictness, 神煞 soft-warn rendering, multi-sentence interleaving, edge cases).
- Frontend: no immediate change. The soft-warn suffix `(注：未在本年算法 evidence 中识别)` is plain text inline, so existing renderers handle it.
- API shape: no breaking change. `narrative` strings may now contain inline suffix markers when 神煞 is unattested; consumers that parse narratives should remain robust.
- Token cost: no change (no prompt edits).
- Algorithm version: increment to `v3-narrative-graceful` so analytics can separate Phase 1 from Phase 2 cohorts.

## Out of Scope

- The retry-on-violation strategy (AI re-prompted with offending term forbidden). Reserved as Phase 3 if sentence-level + soft-warn still leaves a measurable gap.
- The original Phase 2 plan items (applyPolarity priority swap, 神煞 demoted to bracket notes, getYongshenBaseline expansion). Phase 1 already delivered the empty-year improvement those were meant to achieve; further algorithm change is not currently justified.
- Cache migration of v1/v2 rows. Phase 2 only affects newly generated narratives.

## Observability Plan (2 weeks post-launch)

- Empty-year rate: target < 1% (down from current 3.3%).
- Soft-warn frequency: count of "soft_warn" events per dayun call. If < 0.3 per call on average, soft-warn is working as designed (most narratives have no decorative 神煞 issues).
- Hard-set drop rate: count of sentences cleared for 伏吟/反吟/三合/三会/大运合化/用神位等. Aim for ≤ 1 per dayun call (current: ~1 per call, expected to stay flat since Phase 2 doesn't change AI behavior).
- Reader complaints about "(注：未在本年算法 evidence 中识别)" suffix appearing too often.
