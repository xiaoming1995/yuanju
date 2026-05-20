## Context

Prompts 在本项目里同时被两种生命周期管理：

1. **代码生命周期**：每次代码迭代（新增结构化字段、调整问题分支、变 schema）会改 `xxxPromptFallback()` 函数体。代码评审、单元测试都基于"代码 fallback 是当前权威版本"假设。
2. **运营生命周期**：admin 通过 UI 微调措辞、加业务约束、按客户反馈打补丁。这些编辑落在 DB `ai_prompts` 表里，**优先级高于代码 fallback**。

当前服务读取顺序（`compatibility_service.go:303-310`）：

```go
promptConfig, err := repository.GetPromptByModule("compatibility")
tplContent := compatibilityPromptFallback()
if promptConfig != nil && strings.TrimSpace(promptConfig.Content) != "" {
    tplContent = promptConfig.Content  // ← DB 完全 shadow 代码
}
```

这套语义对于"admin 调一句话"场景是对的；但对于"代码 schema 整体演化"场景是错的——admin 一次旧编辑就会无声把所有 schema 升级全部回退。本提案的目标是在保留 admin 编辑能力的同时，给代码一条"我比 DB 新"的可发现通道。

## Goals / Non-Goals

**Goals：**
- 启动期对齐：每次后端启动检查 canonical 注册表 vs DB，对齐未自定义的 prompt 到代码最新版。
- 不丢 admin 编辑：被 admin 改过的 prompt 永不被自动覆盖；admin 看见"已自定义"徽标知道自己在偏离基准。
- 一键回归：admin 想撤销自己的修改时，有"重置为系统默认"按钮，单击即从 canonical 重写。
- 可观察：每个 prompt 行知道自己 `version`（对齐到的代码版本）、`canonical_hash`（对齐时刻的内容指纹）、`is_customized`（是否被人工偏离）。

**Non-Goals：**
- 不做 prompt 多版本并存（前端切换 v1/v2 试用）。
- 不做 prompt diff 可视化（只在徽标提示"已偏离 canonical v{n}"，diff 看 git history）。
- 不接管现有 admin 编辑接口的 RBAC 或审计——已有就用，不改。

## Architecture

```
启动序列
─────────
main.go
  ├─ database.Connect()
  ├─ Migrate(ModeStartup)          ← 现有：跑迁移
  ├─ prompt.SyncCanonical(database.DB)  ← 新增：对齐 canonical
  └─ router.Run(...)

prompt 包
─────────
pkg/prompt/
  canonical.go     ← var Canonical map[string]Definition
                       每个模块持有 {Version, Description, Content, Hash}
                       Hash = sha256(Content) 启动时一次性计算
  sync.go          ← SyncCanonical(db) 主流程
                       for module, def := range Canonical:
                         row := repository.GetPromptByModule(module)
                         switch:
                           row == nil          → repository.InsertCanonical(def)
                           row.is_customized   → log + skip
                           row.version == def.Version → skip (already aligned)
                           else                → repository.UpdateCanonical(def)
  sync_test.go     ← 4 case：insert / skip-customized / update-stale / no-op-aligned

服务层（compatibility_service.go 改造示例）
─────────
- promptConfig := repository.GetPromptByModule("compatibility")
- tplContent := compatibilityPromptFallback()  ← 删
- if promptConfig != nil && ... { tplContent = promptConfig.Content }
+ tplContent := promptConfig.Content  ← SyncCanonical 已保证 DB 有最新内容
+                                       未来调用 prompt.MustGetCanonical("compatibility")
+                                       作为代码侧权威 fallback，仅在 DB 完全无行时使用
```

## Key Decisions

### Decision 1: canonical version 是 free-form string，不是 semver

**Choice：** `Version` 字段用 `"v3-question-aware"` 或 `"2026-05-20-decision-first"` 这种描述性字符串，不强求 SemVer。

**Rationale：** prompt 不是 API、没有兼容性"度"。版本字段唯一作用是"这次代码改了 prompt → 需要对齐 → 用版本号判断是否需要重写"。比较时仅做字符串相等判定（不解析、不排序）。

**Alternative：** SemVer (`1.2.0`) 配 `>` 比较。**Rejected**：prompt 改动不存在 minor/major 语义；强制 SemVer 只是噪音。

### Decision 2: 首次迁移把所有现有行标记 `is_customized=true`

**Choice：** Migration 00011 加列时，UPDATE 所有现存行设 `is_customized=true, version='unversioned'`。

**Rationale：** 保守优先。我们无法从 DB 推断哪些 prompt 是 admin 编辑过的、哪些是初始 seed 留存。一律按"admin 编辑过"对待，避免线上 prompt 内容意外丢失。后果是 admin 需要主动点"重置"才能拉到 canonical 版本——这在 admin UI 加个明显徽标即可解决。

**Alternative：** 启动时用 hash 比对，把 hash==当前 seed hash 的行视为未自定义。**Rejected**：需要记录"baseline 的 hash"，复杂；且如果 admin 改了一个字符又改回来，hash 相同也无法识别。保守标记更安全。

### Decision 3: Reset 端点立即重写，不等下次启动

**Choice：** `POST /api/admin/prompts/:module/reset` handler 直接调用 `repository.ResetPromptToCanonical(...)` 写入 canonical 版本并设 `is_customized=false`。

**Rationale：** admin 操作需要即时反馈。如果只是"标记 is_customized=false 等下次启动"，admin 看不到效果，下一个 AI 调用还是用 admin 的旧版本。

**Alternative：** 重启服务才能生效。**Rejected**：用户体验差，且加重运维负担。

### Decision 4: `compatibility` 模块作为本提案的"首战"

**Choice：** 不在本提案里改其他模块的 prompt（`liunian`、`past_events`、`kb_*`）。仅把 `compatibility` 的 fallback 迁入 canonical 注册表。其他模块保持 `is_customized=true` 状态，admin 现有编辑保留，未来代码侧迁入再说。

**Rationale：** YAGNI + 风险分散。模块迁入需要双侧人工核对（代码 fallback 是否真的比 DB 新、admin 是否有未沉淀的有用调整）。一次性迁全部模块会放大代码评审复杂度，也容易丢 admin 工作。`compatibility` 是本次 5.5 眼测暴露的 bug 现场，作为首战刚好。

**Alternative：** 全模块一次性迁。**Rejected**：上述风险。

### Decision 5: canonical_hash 用 sha256 不是 md5

**Choice：** `canonical_hash CHAR(64)` 存 sha256(content) 的 hex。

**Rationale：** sha256 在 Go 标准库里更"现代"，避免 MD5/SHA1 在审计扫描里被标记为弱算法。64 字符固定长度对索引/对比都很友好。Hash 唯一用途是 sync 时快速判断"DB 内容是否还等于代码当时 INSERT 的内容"——如果 admin 改了内容但忘了改 is_customized（理论上不会，但防御一下），hash 不等也能识别。

## Risks

| 风险 | 缓解 |
|------|------|
| 启动时 SyncCanonical 拿不到 DB（连接还没就绪） | 在 `Migrate()` 之后调用，此时 DB 必然可用；SyncCanonical 内部失败仅记 log 不 panic，服务仍可启动并继续使用旧 prompt |
| 代码 canonical 写错（如缺占位符 `{{.SelfLabel}}`），上线后所有未自定义行被覆盖成坏 prompt | 注册表用例必须含验证测试（占位符存在、模板可解析）；代码评审强制要求改 `canonical.go` 时同步加测试 |
| 误把 admin 编辑的行标记成未自定义，下次启动被覆盖 | 仅在 admin 点"重置"按钮时改 `is_customized=false`；保存编辑的 handler 显式置 `is_customized=true`，二次确认不让漏 |
| `canonical.go` 文件膨胀（~10 模块 × 几百行 prompt = 数千行） | 拆分子文件 `canonical_compatibility.go` / `canonical_past_events.go` 等，按需切分；每个文件只导出 `init()` 中调用 `Register(module, Definition{...})` |
| 多实例部署（K8s 滚动更新）期间，新旧版本同时跑 sync | sync 用 `ON CONFLICT (module) DO UPDATE` + WHERE 条件确保只有自己版本号更新时才覆盖；旧实例不会回退 |

## Migration Path

1. 部署带本变更的版本。Migration 00011 加 3 列、把所有现有行标记 `is_customized=true`。
2. 启动 sync：`compatibility` 模块（之前 DELETE 留空）被 INSERT canonical 版本。其他模块（`liunian`/`past_events`/`kb_*`）的 `is_customized=true`，跳过。
3. Admin 看到 prompt 列表里 `compatibility` 显示"已对齐 v..."徽标，其他显示"已自定义"徽标。
4. 一周观察期：检查后续 AI 调用是否按新 schema 输出（`question_focus.title` 非空）。
5. 后续代码改 `compatibility` prompt → 改 `canonical.go::Canonical["compatibility"].Version` → 部署后启动 sync 自动覆盖到新版（除非 admin 把该模块标记成 customized）。

## Test Strategy

- **`pkg/prompt/sync_test.go`** 4 case：
  - `TestSyncCanonical_InsertsMissingModule`：DB 空，canonical 注册了 1 项 → 同步后 DB 有该行，`is_customized=false`，`version` 匹配。
  - `TestSyncCanonical_SkipsCustomizedRow`：DB 有 `is_customized=true` 旧版 → 同步后内容/版本不变。
  - `TestSyncCanonical_UpgradesStaleAlignedRow`：DB 有 `is_customized=false` 旧版本号 → 同步后内容/版本 = canonical。
  - `TestSyncCanonical_NoOpOnAlignedRow`：DB 有 `is_customized=false` 版本号已等于 canonical → 同步无 UPDATE。
- **`internal/handler/admin_prompt_test.go`** 新增：
  - `TestAdminSavePrompt_SetsCustomizedTrue`：保存后查 DB `is_customized=true`。
  - `TestResetPromptToCanonical_RewritesAndUnflags`：先有 customized 行 → reset → 内容回到 canonical、`is_customized=false`、`version` 匹配。
- **集成验证**：本地 `docker-compose up` 启动，DB 中 `ai_prompts.compatibility` 行 `is_customized=false`、`version='v3-question-aware'`、`canonical_hash` 与 `pkg/prompt/canonical.go::Canonical["compatibility"]` 的 hash 一致。AI 报告 `content_structured.question_focus.title='婚姻适配判断'`（或类似），即与今天手动 DELETE+regenerate 后看到的同分布。

## Observability

- 启动 log：`[prompt-sync] module=compatibility action=insert version=v3-question-aware hash=ab12...`
- 启动 log（跳过场景）：`[prompt-sync] module=past_events action=skip reason=is_customized version=unversioned`
- Admin UI：每个 prompt 卡片右上角徽标
  - 绿色 "已对齐 v3-question-aware"（is_customized=false）
  - 橙色 "已自定义（基准 v3-question-aware）"（is_customized=true, version != 'unversioned'）
  - 灰色 "历史遗留"（version='unversioned'，未走过 canonical 流程）
