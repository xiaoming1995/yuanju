## Context

After Phase 1 (`yongshen-priority-realignment`) shipped, end-to-end measurement on chart 17129161 showed:

```
Baseline (no Phase 1):  60% empty cards
After Phase 1:           3.3% empty cards (18x reduction)
After Phase 1 + Patch 4: 10% empty cards  ← Patch 4 reverted
                          (allowed_keywords prompt-side whitelist
                           backfired — AI read the list as suggestions)
```

The remaining 3.3% breaks down to:
- ~70% are AI writing "伏吟" when the year's evidence has no 伏吟 trigger (relation word)
- ~30% are AI writing a 神煞 name that didn't trigger for this year (decorative usage)

In both cases, `ValidateYearNarrative` currently clears the **entire** narrative (150 chars → 0 chars). The validator gives up too much.

The Phase 1 design.md "Lessons Learned" section already records why the prompt-side whitelist failed; this proposal acts on the conclusion that the right intervention is **validator-side**.

## Goals / Non-Goals

**Goals:**
- Bring empty-card rate from 3.3% toward < 1%.
- Differentiate validator strictness by keyword class so casual 神煞 over-confidence does not destroy structurally sound narratives.
- Keep the validator deterministic, fast, and free of LLM round trips.
- Maintain backward compatibility for callers reading `narrative` field as plain text.

**Non-Goals:**
- Do not retry the AI when validation fails. Retry is a Phase 3 candidate.
- Do not change prompts. Phase 1's prompt is the equilibrium.
- Do not invalidate existing cached narratives. Phase 2 affects newly generated narratives only.
- Do not change the 33-term `validatedKeywords` list itself.

## Decisions

### Decision 1: Sentence boundaries via Chinese punctuation

Use `。!?；…` as primary sentence terminators, plus the full-width counterparts `。！？；……`. Soft-segment on ASCII-only `;.` is not enough — Chinese narratives use full-width.

```go
var sentenceTerminators = []rune{'。', '！', '？', '；', '…', '!', '?', '.', ';'}
```

A sentence is the substring from after the previous terminator (or start) up to and including the next terminator. The terminator stays attached to the sentence (it's part of the preserved fragment).

Edge case: trailing text without a terminator (e.g., "本年大局可观") still counts as a sentence.

**Why this approach**: Chinese narrative text from AI is dense; using paragraph-level granularity would lose too much, and word-level granularity would risk grammatical corruption. Sentence is the right unit — natural句号-bounded clause where removing one leaves the rest grammatically intact.

### Decision 2: Bifurcated keyword strictness

Reclassify `validatedKeywords` (33 items) into 4 buckets with different responses:

```
HardRelation (5)     → Sentence cleared on violation, no marker.
  伏吟 反吟 三合 三会 大运合化

PositionAnchor (3)   → Sentence cleared on violation, no marker.
  用神位 忌神位 喜神位

EventMarker (4)      → Sentence cleared on violation, no marker.
  受冲 受刑 双重命中 力度倍增

Shensha (24)         → Sentence kept with inline marker
                       "(注：未在本年算法 evidence 中识别)"
                       appended after the offending term, or at
                       sentence end if the term recurs in the sentence.
  驿马 桃花 华盖 白虎 丧门 吊客 灾煞 流霞
  天医 天喜 天乙 天德 月德 文昌 太极 福星
  红艳 孤辰 寡宿 羊刃 亡神 劫煞 披麻 咸池
  勾绞 国印
```

**Why this asymmetry**:
- HardRelation/PositionAnchor/EventMarker terms make falsifiable algorithmic claims. AI saying "本年伏吟" when there's no 伏吟 misleads about chart structure. Discarding the sentence is the right call.
- 神煞 are 《滴天髓》-classified as "末" (peripheral). A decorative reference to a non-triggered 神煞 is not an algorithmic falsehood — more a stylistic flourish. Soft-warn lets the user see the content with a transparency annotation.

### Decision 3: ValidateYearNarrative return type

Change from `(bool, string)` to a structured result so the caller can act on per-sentence outcomes:

```go
type NarrativeValidationResult struct {
    // CleanedNarrative is the post-processing narrative text:
    //  • sentences with HardRelation/Position/EventMarker violations removed
    //  • sentences with Shensha violations annotated with inline marker
    // Equal to the input when no violations occur.
    CleanedNarrative string

    // ClearedSentences captures the original sentences that were dropped
    // (after sentence segmentation), one per cleared sentence, for audit.
    ClearedSentences []string

    // SoftWarnedTerms captures the 神煞 names that got the inline marker.
    SoftWarnedTerms []string

    // HardKeywordsHit captures the HardRelation/Position/EventMarker terms
    // that triggered sentence clears (for structured logging).
    HardKeywordsHit []string
}

func ValidateYearNarrative(narrative string, signals []bazi.EventSignal) NarrativeValidationResult
```

The old `bool` "all-or-nothing pass" semantics is gone. Callers compare `CleanedNarrative` to the input or inspect `ClearedSentences` to know whether anything changed.

### Decision 4: Caller migration in GenerateDayunSummariesStream

Replace:
```go
if ok, reason := ValidateYearNarrative(y.Narrative, ...); !ok {
    validatedYears[i].Narrative = ""
}
```

With:
```go
result := ValidateYearNarrative(y.Narrative, ...)
validatedYears[i].Narrative = result.CleanedNarrative
if len(result.ClearedSentences) > 0 || len(result.SoftWarnedTerms) > 0 {
    log.Printf("[ValidatorAction] dayun=%d year=%d cleared=%v soft_warned=%v",
        dy.Index, y.Year, result.HardKeywordsHit, result.SoftWarnedTerms)
}
```

No SSE protocol change — the year card still has a `narrative` field, just possibly with content where it would have been blank before.

### Decision 5: Algorithm version bump to v3-narrative-graceful

Increment `repository.CurrentAlgorithmVersion` from `v2-yongshen-shishen` to `v3-narrative-graceful`. New rows are stamped accordingly; v2 rows remain untouched. Analytics can split Phase 1 vs Phase 2 cohorts cleanly.

## Risks

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Sentence segmentation cuts incorrectly (e.g., quoted text contains 。) | Low | Sentences from AI are descriptive, not dialogue. If we observe it, narrow segmenter to "outside quotes" rule. |
| Inline marker "(注：...)" appears mid-sentence and reads awkwardly | Medium | Place marker after the offending term, not at sentence end. If user feedback is negative, change to footnote-style suffix at sentence end. |
| Bifurcated strictness misses an important new keyword | Low | The 33-term list is unchanged; only the response per term changes. Future additions go through normal validation expansion. |
| Sentence clear removes the only meaningful sentence | Low | If `CleanedNarrative` ends up empty or < 30 chars, fall back to current chip-only rendering on the frontend (already supported). |

## Architecture Diagram

```
┌───────────────────────────────────────────────────────────────────┐
│                AI narrative (150 chars, 4 sentences)              │
└─────────────────────────────┬─────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  Segment by 。!?；…           │
              └───────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  For each sentence, check     │
              │  against 4 keyword buckets    │
              └───────────────────────────────┘
                              │
                ┌─────────────┼─────────────┐
                ▼             ▼             ▼
      ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
      │ Hard /       │ │ 神煞         │ │ No violation │
      │ Position /   │ │ unattested   │ │              │
      │ EventMarker  │ │              │ │              │
      │ unattested   │ │              │ │              │
      ├──────────────┤ ├──────────────┤ ├──────────────┤
      │ CLEAR        │ │ INLINE TAG   │ │ KEEP         │
      │ sentence     │ │ "(注：未在   │ │ sentence     │
      │ (drop)       │ │  evidence)" │ │              │
      └──────────────┘ └──────────────┘ └──────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  Re-join surviving sentences  │
              │  → CleanedNarrative           │
              └───────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │  Persist to ai_dayun_summaries│
              │  algorithm_version = 'v3'     │
              └───────────────────────────────┘
```

## Open Questions

- Q: When the entire narrative is cleared (all 4 sentences hit HardRelation), should we fall back to a structured chip-only render or write `""`?
  - A: Write `""` and let the frontend's existing chip fallback handle it. Same UX as today's empty-year case but rarer.

- Q: Should the inline marker be the same wording for all 24 神煞, or customized?
  - A: Same wording for Phase 2. Customization (e.g., distinguishing "the algorithm has a different 神煞 instead" from "no 神煞 at all") is Phase 3 nice-to-have.

- Q: How is structured logging consumed downstream?
  - A: Phase 2 only emits the log line. Aggregating it into a dashboard is left to the existing token-usage observability work.
