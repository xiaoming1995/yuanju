## Why

The past-events year narrative engine has a structural priority mismatch with classical 八字 doctrine (per 《滴天髓》 任铁樵 注):

- **Algorithm output**: 调候用神 + 扶抑用神 are computed correctly (`yongshen.go::inferYongshenWithTiaohouPriority`) but their downstream usage is shallow. The current `getYongshenBaseline` only checks if the 流年 stem 五行 matches `Yongshen` / `Jishen` strings — it never exposes which 十神 are favorable/adverse, and the AI prompt does not receive any explicit 喜忌十神 directive.
- **AI behavior**: With no explicit 喜忌十神 list, the AI fills 流年 narratives with high-frequency 命理 vocabulary (驿马, 羊刃, 白虎, 伏吟) regardless of whether the algorithm actually triggered them.
- **Validator side-effect**: `ValidateYearNarrative` then strips the whole narrative when a keyword has no evidence source, producing empty year cards.

Real impact today (2026-05-19 measurement):
- 60% of cached dayun summaries have ≥1 empty year narrative.
- One dayun has 5/10 years empty.
- Recent 5 validator drops were all "AI wrote 驿马/伏吟/羊刃/白虎 but evidence has none".

This change calibrates the year-judgment pipeline so 调候/扶抑用神 → 喜忌五行 → 喜忌十神 becomes the **primary judgment axis** that the AI sees, and 神煞 / 三合三会 become secondary modifiers — aligning the algorithm with the 滴天髓 doctrine "五行生克为本，神煞为末".

## What Changes

Phase 1 (this proposal — low risk, observable, fully reversible):

- Compute explicit **喜十神 / 忌十神** lists for each `BaziResult` from `Yongshen` + `Jishen` + `strengthLevel` using the 古法 mapping (身旺 → 喜泄克耗；身弱 → 喜生扶；中和 → 弱判定，以调候为主).
- Surface a confidence band (`vstrong`/`vweak` = hard; `strong`/`weak` = medium; `neutral` = soft) so the AI does not force a binary judgment on 中和 命主.
- Inject the favorable/adverse 十神 list (with confidence) into the `GenerateDayunSummariesStream` prompt template so the AI can directly reference it when writing each year card.
- Add `algorithm_version` column to `ai_dayun_summaries` and `ai_reports` so future algorithm changes can be tracked, and old cached reports can display a version badge.
- Test coverage: unit tests for the mapping function, regression test that 流年 narratives no longer fabricate 十神 polarities, integration test that the new prompt fields appear in the rendered template.

Phase 2 (deferred — depends on Phase 1 metrics, separately proposed):

- Realign `applyPolarity` so 神煞 no longer overrides 用神 baseline.
- Enhance `getYongshenBaseline` to consider 地支藏干透出十神 + 喜忌神位 受冲克合 interactions.
- Demote 20+ minor 神煞 from independent `EventSignal` to bracket notes inside main-line evidence (keeping 用神位/忌神位/喜神位, 伏吟/反吟, 羊刃/白虎/天乙/文昌 as independent signals).
- Migrate stale cache rows with empty years.

Out of scope:

- 从化局 (从势/从财/从官) judgment branch — a separate large project; ~2% of charts but ~0% currently correct. Track separately.
- 实战派 大运修正 of 喜忌十神 — adds 8x complexity for ~10% accuracy gain on edge cases; keep 古法 first.

## Capabilities

### New Capabilities
- `yongshen-driven-flow-year-judgment`: Explicit 喜忌十神 derivation from 调候/扶抑用神, surfaced via prompt for AI year-card generation.

### Modified Capabilities
- None in Phase 1. Phase 2 will modify `past-events-year-narrative-depth` if enacted.

## Impact

- Backend: `pkg/bazi/yongshen.go` adds `BuildFavorableShishen()`. `pkg/bazi/engine.go::BaziResult` gains `FavorableShishen`, `AdverseShishen`, `ShishenConfidence` fields. `internal/service/report_service.go::GenerateDayunSummariesStream` updates prompt template. `pkg/database/migrations/` gains migration for `algorithm_version` columns.
- Backend tests: `pkg/bazi/yongshen_test.go` adds mapping cases for 5 strength tiers × 4 day-master 五行. New integration test in `internal/service/` for prompt template rendering.
- Frontend: no immediate change. Future enhancement (out of scope) could surface 喜忌十神 on the chart page.
- API shape: no breaking change. `BaziResult` JSON gains 3 new fields (additive).
- Cost and performance: prompt grows by ~1 short line (~30 tokens / dayun call → ~240 tokens / full past-events recap → ~¥0.001 per recap at deepseek-v4-pro pricing). Negligible.
- Database migration: additive only (`algorithm_version` columns nullable, default v1).
- Backwards compatibility: existing reports keep working. New reports tagged `algorithm_version='v2-yongshen-shishen'`.

## Observability Plan (2 weeks post-launch)

Before deciding whether to proceed with Phase 2, measure:

- Empty year rate: target <30% (down from current 60%).
- Validator drop rate per dayun call: target <0.5 (down from current ~1.0).
- AI 神煞 fabrication rate: count "narrative has 驿马/羊刃/白虎/伏吟 not in evidence" in token_usage_logs samples.
- Token cost: confirm prompt growth stays under +5% of total recap cost.
