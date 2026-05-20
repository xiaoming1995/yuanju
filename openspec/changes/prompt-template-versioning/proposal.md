## Why

The system has two sources of truth for AI prompts that silently disagree:

- **代码 fallback**：每个模块（`compatibility`、`past_events`、`liunian` 等）在 `internal/service/*.go` 中维护一个 `xxxPromptFallback()` 函数，包含与代码期望输出 schema 严格对齐的 prompt（`question_focus`/`decision_advice`/`stage_risks` 等结构化字段）。
- **DB 表 `ai_prompts`**：可通过 admin UI 编辑；服务层优先级 `DB > 代码 fallback`（`compatibility_service.go:303-310`）。

实际事故（2026-05-20 5.5 眼测发现）：
- 代码层 `compatibilityPromptFallback` 已加 `question_focus`、`relationship_diagnosis`、`decision_advice`、`{{.RelationshipStageLabel}}`、`{{.PrimaryQuestionLabel}}` 等新字段；
- DB 中的 `ai_prompts.compatibility` 仍是旧版（811 字符，零新字段），来自不知何时的 admin 编辑或老 migration；
- 服务层走 DB → AI 收到旧 prompt → 按旧 schema 输出 → `question_focus`/`decision_advice` 全空；
- 修复只能 `DELETE FROM ai_prompts WHERE module='compatibility'`，但任何后续 admin 编辑都会重新覆盖代码 fallback，bug 回归。

根因：代码 prompt 隐含版本演化，但 DB 行不持有"对齐到哪一版代码"的信息，sync 链路不存在。

## What Changes

1. **结构层**：`ai_prompts` 表新增 `version`（VARCHAR）+ `is_customized`（BOOLEAN）+ `canonical_hash`（CHAR(64)）三列。
2. **代码层**：新增包 `pkg/prompt`，定义 `Canonical map[string]Definition{Version, Content, Description}` 注册表。把现有各模块的 `xxxPromptFallback()` 内容迁入注册表。
3. **同步层**：`pkg/prompt.SyncCanonical(db)` —— 启动时执行，按模块逐项决策：
   - DB 无该模块 → INSERT 代码版本（`is_customized=false`）
   - DB 有该模块 且 `is_customized=false` 且版本号低于代码 → UPDATE 到代码版本
   - DB 有该模块 且 `is_customized=true` → 跳过（admin 主导），但记录 `canonical_hash`/`version` 用于日后 diff 展示
4. **Admin 编辑链**：admin UI 保存 prompt 时自动置 `is_customized=true`。
5. **重置链**：新增 `POST /api/admin/prompts/:module/reset` 端点 → `is_customized=false` 且立即重写为当前代码 canonical 版本。前端 admin prompt 编辑页加"重置为系统默认"按钮 + "已自定义/对齐 v{ver}"状态徽标。
6. **首次迁移**：所有已存在的 prompt 行 `is_customized=true`、`version='unversioned'`（保守—不丢失任何 admin 已做的编辑）。`compatibility` 模块行已被本次眼测时手动 DELETE，启动时由 SyncCanonical 重新 INSERT 为 canonical 版本。

## Capabilities

### New Capabilities
- `prompt-template-versioning`：代码侧维护 prompt 规范版本注册表 + DB 启动期对齐 + admin 自定义保护 + 一键重置。

### Modified Capabilities
- None.

## Impact

- **Backend**：
  - 新增 `backend/pkg/prompt/` 包（`canonical.go` 注册表 + `sync.go` 对齐器 + 测试）。
  - 1 个 migration（`00011_ai_prompt_versioning.sql`）：加 3 列 + 把现有行标记 `is_customized=true`。
  - `cmd/api/main.go` 启动序列加 `prompt.SyncCanonical(database.DB)`，紧随 `Migrate(ModeStartup)` 之后。
  - `internal/service/compatibility_service.go::compatibilityPromptFallback`（740 行文件第 602 行起）迁移到 `pkg/prompt/canonical.go::Canonical["compatibility"]`，service 调用 `prompt.GetCanonical("compatibility")` 取兜底。
  - `internal/handler/admin_prompt.go`：保存 prompt 时设 `is_customized=true`；新增 `ResetPromptToCanonical` handler。
  - `internal/repository/prompt_repository.go`：新增 `ResetPromptToCanonical(module, version, content, hash)`、`SetCustomized(module, bool)`。
- **Frontend**：admin prompt 编辑页（`src/pages/admin/PromptSettings.tsx`）加"已对齐 v{ver}"/"已自定义"徽标 + 单击重置按钮 + 二次确认。
- **API**：新增 `POST /api/admin/prompts/:module/reset`；现有 `PUT /api/admin/prompts/:module` 隐式置 `is_customized=true`（不破坏兼容）。
- **Database migration**：additive only（3 列均带 default，已有行被填充为保守值）。
- **Cost/perf**：启动期一次 SyncCanonical，约 N 次 SELECT + 至多 N 次 UPSERT（N=模块数≈10）；运行期零开销。
- **Backwards compat**：现有所有 prompt 行被标记 `is_customized=true`，启动 sync 默认不动它们，admin 后续编辑也不会被覆盖。唯一被 sync INSERT 的是 `compatibility` 模块（因本次手动 DELETE 留下空位）—— 与本次眼测的预期修复一致。

## Out of Scope

- 多语言 prompt（cn/en 双轨）。本次只统一中文 prompt 版本治理。
- Prompt A/B 实验框架。canonical 仅维护"最新稳定版"。
- 历史版本回滚 UI。如需回退，admin 仍可手动编辑或重置后再调。
- Prompt 内容自动验证（如必含 `{{.SelfLabel}}` 占位符）。canonical 写错由代码评审兜底。
