## 1. Schema 与 Migration

- [ ] 1.1 创建 `backend/pkg/database/migrations/00011_ai_prompt_versioning.sql`：`ALTER TABLE ai_prompts ADD COLUMN version VARCHAR(64) NOT NULL DEFAULT 'unversioned'`、`ADD COLUMN is_customized BOOLEAN NOT NULL DEFAULT FALSE`、`ADD COLUMN canonical_hash CHAR(64) NOT NULL DEFAULT ''`。
- [ ] 1.2 同 migration 末尾追加 `UPDATE ai_prompts SET is_customized = TRUE` 把所有现存行标记为 admin 自定义（保守策略，避免线上 prompt 被首次启动覆盖）。
- [ ] 1.3 启动 `--migrate-dry-run` 确认 SQL 语法、`--migrate-apply` 实测在本地 dev DB 跑通。

## 2. `pkg/prompt` 注册表

- [ ] 2.1 创建 `backend/pkg/prompt/canonical.go`：定义 `type Definition struct { Version, Description, Content string; Hash string }` + `var Canonical = map[string]Definition{}` + `func Register(module string, def Definition)` + `func MustGet(module string) Definition`。
- [ ] 2.2 在 `Register` 内对 `def.Content` 算 sha256 写入 `def.Hash`，避免每次 sync 重新算。
- [ ] 2.3 创建 `backend/pkg/prompt/canonical_compatibility.go`：在 `init()` 中调用 `Register("compatibility", Definition{Version: "v3-question-aware", Content: ..., Description: "合盘决策咨询 prompt（含 question_focus / decision_advice / stage_risks）"})`，Content 从 `internal/service/compatibility_service.go::compatibilityPromptFallback` 函数体迁移过来。
- [ ] 2.4 单测 `backend/pkg/prompt/canonical_test.go`：断言 `MustGet("compatibility")` 返回非空 Content、Version 匹配、Hash 长度=64；断言 `MustGet("unknown")` panic。

## 3. Sync 主流程

- [ ] 3.1 创建 `backend/pkg/prompt/sync.go::SyncCanonical(db *sql.DB) error`，遍历 `Canonical`：对每个 module 调 `repository.GetPromptByModule(module)`，按"DB 无行 / is_customized=true / version 匹配 / 其他"四分支决策。
- [ ] 3.2 在 `repository.prompt_repository.go` 新增 `InsertCanonical(module, version, content, hash, description string) error`、`UpdateCanonicalContent(module, version, content, hash string) error`、`SetCustomized(module string, customized bool) error`、`ResetToCanonical(module, version, content, hash string) error`。
- [ ] 3.3 SyncCanonical 失败不 panic，仅 `log.Printf("[prompt-sync] module=%s action=skip reason=db_error err=%v", module, err)`；服务仍能启动。
- [ ] 3.4 单测 `backend/pkg/prompt/sync_test.go` 四 case：missing→insert / customized→skip / stale→upgrade / aligned→noop。使用本地 testdb 或 mock。

## 4. main.go 启动序列接入

- [ ] 4.1 `backend/cmd/api/main.go` 在 `database.Migrate(database.ModeStartup)` 调用之后、`router.Run(...)` 之前加 `prompt.SyncCanonical(database.DB)`。
- [ ] 4.2 启动日志增加 `log.Printf("[prompt-sync] module=%s action=%s version=%s hash=%s...", ...)` 每模块一行，便于运维确认。

## 5. Service 层迁移

- [ ] 5.1 删除 `backend/internal/service/compatibility_service.go::compatibilityPromptFallback`（约 line 602-740）；改为 `prompt.MustGet("compatibility").Content`。
- [ ] 5.2 改 `compatibility_service.go:303-310` 读取顺序：`promptConfig := repository.GetPromptByModule("compatibility"); if promptConfig != nil && strings.TrimSpace(promptConfig.Content) != "" { tplContent = promptConfig.Content } else { tplContent = prompt.MustGet("compatibility").Content }`。这样 SyncCanonical 成功的常态走 DB（已含 canonical 内容），万一 sync 失败 + DB 为空时仍可代码兜底。
- [ ] 5.3 更新 `backend/internal/service/compatibility_service_test.go::TestCompatibilityPromptFallback_*`：原本断言函数返回字符串，改为断言 `prompt.MustGet("compatibility")` 返回的 Content 包含相同模板变量与问题分支文案。
- [ ] 5.4 验证 `go test ./internal/service/... ./pkg/prompt/...` 全过。

## 6. Admin 编辑链

- [ ] 6.1 `backend/internal/handler/admin_prompt.go` 中 PUT prompt handler：保存成功后调 `repository.SetCustomized(module, true)`。
- [ ] 6.2 新增 `ResetPromptToCanonical(c *gin.Context)` handler：从 URL 取 `module`，从 `prompt.MustGet(module)` 取 canonical，调 `repository.ResetToCanonical(...)`，返回更新后的 prompt。
- [ ] 6.3 `backend/cmd/api/main.go` 注册路由 `admin.POST("/prompts/:module/reset", middleware.AdminAuth(), handler.ResetPromptToCanonical)`。
- [ ] 6.4 handler 测试：`TestAdminSavePrompt_SetsCustomizedTrue` + `TestResetPromptToCanonical_RewritesAndUnflags`。

## 7. Frontend admin UI

- [ ] 7.1 `frontend/src/pages/admin/PromptSettings.tsx`：扩展 prompt 列表项渲染，依据 `is_customized` + `version` 显示三态徽标（绿/橙/灰）。
- [ ] 7.2 同页面加"重置为系统默认"按钮，触发二次确认 modal，确认后调 `POST /api/admin/prompts/:module/reset` 并刷新本地状态。
- [ ] 7.3 `frontend/src/lib/adminApi.ts` 新增 `resetPromptToCanonical(module: string)`。
- [ ] 7.4 静态测试 `frontend/tests/admin-prompt-versioning.test.mjs`：断言文件中含徽标三态文案 + 重置按钮 wiring。

## 8. 集成验证

- [ ] 8.1 本地 `docker-compose down && docker-compose up --build`。
- [ ] 8.2 查 `ai_prompts.compatibility` 行：`is_customized=false`、`version='v3-question-aware'`、`canonical_hash` 长度 64 且与 `pkg/prompt/canonical_compatibility.go` 计算的 hash 一致。
- [ ] 8.3 查其他模块（`liunian`/`past_events`/`kb_*`）：`is_customized=true`、`version='unversioned'`（不变）。
- [ ] 8.4 创建新合盘 reading（带 `primary_question='marriage_suitability'`），生成 AI 报告，断言 `content_structured.question_focus.title` 非空、含"婚姻"或"适配"语义。
- [ ] 8.5 admin UI 打开 prompt 编辑页，确认徽标显示正确。点击 `compatibility` 的"重置"按钮（其本已对齐，应为 no-op 但确认 hash 不变）；点击 `liunian` 的"重置"按钮，刷新后变绿色"已对齐"且 version 字段更新。
- [ ] 8.6 admin UI 编辑 `compatibility` prompt 加一行注释保存，刷新后徽标变橙色"已自定义（基准 v3-question-aware）"。

## 9. 文档

- [ ] 9.1 更新 `CLAUDE.md` 的 "Key Conventions" 段：补充"AI prompts 由 `pkg/prompt/canonical.go` 注册并在启动期对齐到 DB；代码侧更新 prompt 时改 `Version` 字段；admin 编辑会置 is_customized=true 阻止后续 sync 覆盖；admin UI 有'重置'按钮回到 canonical。"
- [ ] 9.2 `backend/pkg/prompt/README.md`：说明何时新增模块、如何 bump version、如何写测试。
