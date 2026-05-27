# Compatibility Scoring Formula v3.1 — Relax Threshold

## Why
v3 评分（合属相 50 + 合纳音 20 + 合日柱 10 + 合八字 20）将「六合/三合」与「纳音相生/相同」设为单一闸门，地支双生（如亥子/寅卯）与五行相生（如子→寅）均得 0。统计上约 73% 的地支两两组合在合属相模块得 0；用户在真实案例 (1996-02-08 20时) vs (1995-02-02 16时) 上拿到 0/100 + 空 evidence 列表，无法理解。

## What Changes
- 合属相：单档 50/0 → 三级 50/30/20/0（双生 30、相生 20）
- 合日柱：三级 0/5/10 → 四级 0/3/5/10（日支同行/相生 → 下档 3）
- 合八字：内部调用更新后的 scoreDayPillar；归一化函数不变
- 合纳音：不变
- `analysis_version`: `"v3"` → `"v3.1"`；v3 旧记录保留不重算
- AI prompt 与 evidence 文案同步更新

## Impact
- Affected specs:
  - **MODIFIED:** `compatibility-scoring-formula`（zodiac / day_pillar / eight_chars / analysis_version 四个 Requirement）
- Affected code:
  - backend/pkg/bazi/compatibility_scoring.go
  - backend/pkg/bazi/compatibility_scoring_test.go
  - backend/pkg/bazi/compatibility_evidence.go
  - backend/internal/service/compatibility_service.go
  - backend/pkg/prompt/canonical_compatibility.go
  - frontend/src/lib/api.ts
  - frontend/src/pages/CompatibilityResultPage.tsx
  - frontend/src/pages/CompatibilityHistoryPage.tsx
- DB migration: 无 schema 变化
- 历史记录: v3 行保留 analysis_version='v3' 不重算
