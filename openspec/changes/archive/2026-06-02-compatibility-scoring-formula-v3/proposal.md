# Compatibility Scoring Formula v3

## Why
合盘模块当前评分算法（11 信号源 × 4 维度 × evidence 加权 + 来源贡献封顶）存在双重计分、十神过粗、五行失衡触发门槛过严、`compatibility.go` 超 1510 行违反 500 行硬约束等问题。用户给出的传统命理 100 分制公式（合属相 50 + 合纳音 20 + 合日柱 10 + 合八字 20）表达更直接，且分模块解释更容易向用户展示。

## What Changes
- 替换 `backend/pkg/bazi/compatibility.go` 全部评分逻辑为「纯加分制」4 模块算法
- `CompatibilityDimensionScores` 字段重命名：`attraction/stability/communication/practicality` → `zodiac/nayin/day_pillar/eight_chars`
- 新增 `overall_score` 字段（INTEGER 列 + JSON 顶层字段）
- `analysis_version` 由 `"v2"` 升至 `"v3"`；v1/v2 旧记录通过前端 version 分支保留渲染
- 旧两份 spec `compatibility-explainable-compatibility-scoring` / `compatibility-depth-signal-engine` 待归档（语义被新公式推翻）
- frontend 历史页 + 结果页 version 分支双渲染

## Impact
- Affected specs:
  - **REMOVED:** `compatibility-explainable-compatibility-scoring`（v3 取消 evidence 加权与来源贡献封顶模型）
  - **REMOVED:** `compatibility-depth-signal-engine`（v3 取消 11 类信号源体系）
  - **ADDED:** `compatibility-scoring-formula`（本 change）
- Affected code:
  - backend/pkg/bazi/compatibility*.go（重写 + 4 个新文件：compatibility_nayin.go / compatibility_scoring.go / compatibility_evidence.go / compatibility_assessment.go）
  - backend/internal/model/compatibility.go（字段重命名）
  - backend/internal/service/compatibility_service.go（version + 映射）
  - backend/internal/repository/compatibility_repository.go（overall_score 列）
  - backend/pkg/database/migrations/00012_compatibility_v3_analysis.sql（新增）
  - backend/pkg/prompt/canonical_compatibility.go（prompt 改造）
  - frontend/src/lib/api.ts（类型联合 v1/v2/v3）
  - frontend/src/lib/compatibilityPersonality.ts（参数缩窄为 Legacy）
  - frontend/src/pages/CompatibilityResultPage.tsx & .css（v3 组件 + 分支）
  - frontend/src/pages/CompatibilityHistoryPage.tsx & .css（version 分支）
- DB migration: 新增列 `overall_score INTEGER NOT NULL DEFAULT 0`
- 历史记录：保留不动；前端按 `analysis_version` 切渲染分支
