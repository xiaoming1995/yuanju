## ADDED Requirements

### Requirement: ai_prompts 行持有版本和自定义状态

`ai_prompts` 表 SHALL 为每行持有 `version` (VARCHAR(64) NOT NULL)、`is_customized` (BOOLEAN NOT NULL)、`canonical_hash` (CHAR(64) NOT NULL) 三个字段，用于区分代码 canonical 对齐状态与 admin 手动编辑状态。

#### Scenario: 启动迁移后，所有现存 prompt 行被保守标记

- **WHEN** 迁移 `00011_ai_prompt_versioning.sql` 执行
- **THEN** 所有现存 `ai_prompts` 行 SHALL 有 `is_customized = TRUE`
- **AND** `version` SHALL 等于 `'unversioned'`
- **AND** `canonical_hash` SHALL 为空字符串
- **AND** 任何 admin 已编辑或初始 seed 的 `content` 字段 SHALL NOT 被改动

#### Scenario: 新插入的 canonical 行带有正确字段

- **WHEN** 启动期 `SyncCanonical` 对某模块执行 INSERT 路径
- **THEN** 该行 SHALL 有 `is_customized = FALSE`
- **AND** `version` SHALL 等于 `pkg/prompt.Canonical[module].Version`
- **AND** `canonical_hash` SHALL 等于 sha256(content) 的 hex 字符串（64 字符）

---

### Requirement: 代码侧 canonical 注册表为权威 prompt 源

The system SHALL 在 `pkg/prompt` 包中维护一个 `Canonical map[string]Definition` 注册表，其中每个 `Definition` 包含 `Version`, `Description`, `Content`, `Hash` 字段。任何模块要被 SyncCanonical 管理 SHALL 在 `init()` 中调用 `Register(module, Definition{...})`。

#### Scenario: 注册时自动计算 hash

- **WHEN** 调用 `prompt.Register("compatibility", Definition{Content: "你是..."})`
- **THEN** 该 Definition 在写入 `Canonical` map 前 SHALL 有 `Hash = sha256(Content)` 的 hex 字符串
- **AND** 后续 `Canonical["compatibility"].Hash` SHALL 直接可读

#### Scenario: 未注册模块 MustGet 抛 panic

- **WHEN** 调用 `prompt.MustGet("not-a-real-module")`
- **THEN** 该调用 SHALL panic 并指出未注册的 module 名
- **AND** 上层调用方 SHALL NOT 静默回退到空字符串

---

### Requirement: 启动期对齐保留 admin 自定义

`SyncCanonical(db *sql.DB) error` SHALL 在后端启动期被 `main.go` 调用一次，紧随 `database.Migrate(ModeStartup)` 之后、`router.Run(...)` 之前。该函数 SHALL 遍历 `Canonical` 中的每个模块并执行以下决策。

#### Scenario: 模块在 DB 中不存在 → 插入 canonical

- **WHEN** `SyncCanonical` 处理模块 M
- **AND** `repository.GetPromptByModule(M)` 返回 `(nil, nil)`
- **THEN** SHALL 执行 `repository.InsertCanonical(M, def.Version, def.Content, def.Hash, def.Description)`
- **AND** 新行 SHALL 有 `is_customized = FALSE`
- **AND** 启动日志 SHALL 包含 `[prompt-sync] module=M action=insert version=...`

#### Scenario: 模块在 DB 中已被 admin 自定义 → 跳过

- **WHEN** `SyncCanonical` 处理模块 M
- **AND** DB 中该行 `is_customized = TRUE`
- **THEN** SHALL 不修改任何字段
- **AND** 启动日志 SHALL 包含 `[prompt-sync] module=M action=skip reason=is_customized`

#### Scenario: 模块在 DB 中版本号已等于 canonical → 跳过

- **WHEN** `SyncCanonical` 处理模块 M
- **AND** DB 中该行 `is_customized = FALSE`
- **AND** DB 中该行 `version` 字段值等于 `Canonical[M].Version`
- **THEN** SHALL 不修改任何字段
- **AND** 启动日志 SHALL 包含 `action=noop`

#### Scenario: 模块在 DB 中是未自定义的旧版本 → 升级

- **WHEN** `SyncCanonical` 处理模块 M
- **AND** DB 中该行 `is_customized = FALSE`
- **AND** DB 中该行 `version` 不等于 `Canonical[M].Version`
- **THEN** SHALL 执行 `repository.UpdateCanonicalContent(M, def.Version, def.Content, def.Hash)`
- **AND** `is_customized` 字段 SHALL 保持 FALSE
- **AND** 启动日志 SHALL 包含 `action=upgrade from=旧版本 to=新版本`

#### Scenario: SyncCanonical 内部失败不阻断服务启动

- **WHEN** `SyncCanonical` 在处理某模块时遇到 DB 错误
- **THEN** SHALL 记录 `log.Printf("[prompt-sync] module=M action=skip reason=db_error err=%v", err)`
- **AND** SHALL 继续处理后续模块
- **AND** SHALL 返回 nil（不让 main.go 因 sync 失败而退出）

---

### Requirement: Admin 编辑自动标记自定义

The admin prompt save endpoint (`PUT /api/admin/prompts/:module`) SHALL 在 prompt 内容被任何修改保存后，自动将该行的 `is_customized` 字段置为 `TRUE`。Admin 用户 SHALL NOT 需要手动设置该 flag。

#### Scenario: 编辑保存后 is_customized 变 TRUE

- **GIVEN** 一个 prompt 行当前 `is_customized = FALSE`
- **WHEN** admin 通过 PUT endpoint 提交新 content
- **THEN** 行的 `content` 字段 SHALL 被更新为新值
- **AND** `is_customized` SHALL 等于 TRUE
- **AND** `version` 字段 SHALL 保持原值（admin 编辑不改 canonical 版本号）

#### Scenario: 编辑后下次 sync 跳过此行

- **GIVEN** 一个 prompt 行刚被 admin 编辑（`is_customized = TRUE`）
- **WHEN** 后端重启并执行 `SyncCanonical`
- **THEN** 该行的 `content` 字段 SHALL NOT 被覆盖
- **AND** 启动日志 SHALL 显示 `action=skip reason=is_customized`

---

### Requirement: 一键重置回 canonical

The system SHALL 提供 `POST /api/admin/prompts/:module/reset` endpoint。该端点 SHALL 立即（不等下次启动）将指定模块的 `content`, `version`, `canonical_hash` 重写为代码注册表当前版本，并将 `is_customized` 置为 FALSE。

#### Scenario: 重置覆盖自定义内容

- **GIVEN** 一个 prompt 行 `is_customized = TRUE`，content 是 admin 自定义文本
- **WHEN** admin 通过 `POST /api/admin/prompts/:module/reset` 重置
- **THEN** 行的 `content` SHALL 等于 `prompt.MustGet(module).Content`
- **AND** `version` SHALL 等于 `prompt.MustGet(module).Version`
- **AND** `canonical_hash` SHALL 等于 `prompt.MustGet(module).Hash`
- **AND** `is_customized` SHALL 等于 FALSE

#### Scenario: 重置未注册模块返回 404

- **WHEN** admin 调用 `POST /api/admin/prompts/not-a-real-module/reset`
- **THEN** 该端点 SHALL 返回 HTTP 404
- **AND** response body SHALL 包含错误信息 `unknown module: not-a-real-module`
- **AND** SHALL NOT panic 也 SHALL NOT 修改 DB

---

### Requirement: Admin UI 三态徽标 + 重置按钮

The admin prompt editing page (`frontend/src/pages/admin/PromptSettings.tsx`) SHALL 为每个 prompt 模块显示一个状态徽标和一个"重置为系统默认"按钮。

#### Scenario: 三态徽标对应数据库状态

- **WHEN** 渲染 prompt 列表项
- **THEN** SHALL 显示徽标如下：
  - `is_customized = FALSE` 且 `version != 'unversioned'` → 绿色徽标 "已对齐 v{version}"
  - `is_customized = TRUE` 且 `version != 'unversioned'` → 橙色徽标 "已自定义（基准 v{version}）"
  - `version = 'unversioned'` → 灰色徽标 "历史遗留"

#### Scenario: 重置按钮二次确认

- **WHEN** admin 单击某 prompt 的"重置为系统默认"按钮
- **THEN** SHALL 弹出二次确认 modal，含"重置将丢弃当前自定义内容，确定继续？"文案
- **AND** 仅当 admin 在 modal 中明确确认后，前端 SHALL 调用 `POST /api/admin/prompts/:module/reset`
- **AND** 响应成功后，列表项徽标 SHALL 立即变成绿色"已对齐"
