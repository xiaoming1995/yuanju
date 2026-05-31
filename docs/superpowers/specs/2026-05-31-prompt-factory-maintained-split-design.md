# Prompt「出厂版 / 维护版」两层模型 + 漂移可见

日期：2026-05-31
状态：设计待评审

## 背景与问题

线上与本地的合盘 AI 报告结构不一致：线上停在旧 prompt（仅 `summary/dimensions/risks/advice`），本地是 v3.1（含 `personality_comparison` / `relationship_strategy` 等）。

排查结论——**与 `reading_id` 缓存无关**，根因在 prompt 的存储与同步机制：

- prompt 有两份：代码侧 canonical（`backend/pkg/prompt/canonical_*.go`）与数据库 `ai_prompts` 表。运行时只读 DB 那份。
- 启动时 `SyncCanonical` 把代码 canonical 同步进 DB，但对 `is_customized=true` 的行**静默跳过**。
- 后台「保存」按钮（`UpdatePrompt`）**只要保存过一次就把 `is_customized` 置 true**（不判断内容是否真改）。
- 结果：线上某次（可能无意的）后台保存给合盘 prompt 上了锁 → 之后 v3→v3.1 的 canonical 升级被启动同步永久跳过 → 线上 AI 一直用旧版。而漂移在后台页面完全不可见（`VersionBadge` 只看 DB 行自身字段，从不与代码当前版本比对）。

## 目标

1. 杜绝「部署悄悄改变线上 AI 输出」与「自定义被无声跳过升级」两类问题。
2. 让 prompt 的版本关系对管理员**可见、可控**。
3. 不引入数据库迁移。

## 核心模型：出厂版 / 维护版

把现在纠缠的关系掰成清晰两层：

```
出厂版（代码 canonical，prompt.Lookup）     维护版（DB ai_prompts 行）
────────────────────────────────         ──────────────────────────
• 只读，运行时从不直接使用                  • 运行时唯一读取源
• 工程师升级它（v3 → v3.1）                 • 仅在管理员主动操作时才变
• 作为「可查看 + 可一键采用」的参考          • 出厂版永不自动覆盖它
```

**两条铁律：**
1. AI 只读维护版（DB）。
2. 出厂版永不自动覆盖维护版；升级靠管理员手动「采用」。

**接受的代价：** 每次 prompt 升级，每个环境都要手动点一次「采用出厂新版」才生效。换来的是「升级前一定可见、可审阅」，且 prompt 改动不再是部署的副作用。

## 漂移判定（读取时实时计算，3 态）

复用现有字段，不加列：
- DB 行的 `version` + `canonical_hash` = 这份维护版**基于哪个出厂版**（分支点）。
- 出厂版当前 `Version` + `Hash` 来自 `prompt.Lookup(module)`。

| drift_status | 条件 | 徽标文案 | 可用操作 |
|---|---|---|---|
| `aligned` | `DB.canonical_hash == factory.Hash` 且 `DB.content == factory.content` | ✅ 已是出厂版 {factory.Version} | 编辑 |
| `customized` | `DB.canonical_hash == factory.Hash` 且 `DB.content != factory.content` | ✏️ 已自定义（基于出厂 {factory.Version}） | 编辑 / 重置为出厂 |
| `outdated` ★ | `DB.canonical_hash != factory.Hash` | ⚠️ 出厂已更新到 {factory.Version}（你基于 {DB.version}） | 采用出厂新版 / 编辑 |
| `unregistered` | 代码无此 canonical | 历史遗留 | — |

`outdated` 就是本次事故第一次变得可见的状态。

## 后端改动

### 1. SyncCanonical：删掉自动覆盖（`backend/pkg/prompt/sync.go`）

决策表简化为：
```
DB 没这行  → InsertCanonical（补种初始值，首次部署需要）
DB 已存在  → noop（永不覆盖，无论是否 is_customized、版本是否一致）
```
即移除现有的「version 不匹配 → UpdateCanonicalContent」覆盖分支与 `is_customized` 跳过分支。`is_customized` 列保留但不再用于同步决策（避免迁移）。

### 2. 保存逻辑改写（`backend/internal/handler/admin_prompt.go` `UpdatePrompt`）

保存时把维护版的**分支点更新到当前出厂版**，并据内容决定 `is_customized`（仅作信息标记）：
```
def = prompt.Lookup(module)
isCustom = sha256(content) != def.Hash
UpdateMaintained(module, content, def.Version, def.Hash, isCustom)
```
含义：保存即「我已对照当前出厂版确认」，故分支点对齐到当前出厂版——`outdated` 在保存后自动消除。顺带堵住「原样保存即上锁」的旧 footgun（存与出厂一字不差的内容 → `is_customized=false`）。

### 3. 新增 repo 函数（`backend/internal/repository/prompt_repository.go`）

`UpdateMaintained(module, content, version, hash string, isCustomized bool)`：一次性写 `content / version / canonical_hash / is_customized`。替代现有「`UpdatePrompt` + `SetCustomized`」两步。

### 4. 列表附加计算字段（`GetPrompts`）

每个返回项附 `canonical_version`、`drift_status`（用 `prompt.Lookup` 与 DB 行比对得出，不入库）。返回用一个 handler 层 DTO，不污染 `model.AIPrompt`。

### 5. 新端点：取出厂版内容

`GET /api/admin/prompts/:module/canonical` → `{version, content}`，供前端「采用出厂新版」把出厂内容载入编辑器。

### 6. 保留不变

`ResetToCanonical` 与 `POST /api/admin/prompts/:module/reset`（硬覆盖回出厂版、清 `is_customized`）继续作为「重置为出厂」入口。

## 前端改动（`frontend/src/pages/admin/PromptSettings.tsx` + `lib/adminApi.ts`）

1. `PromptRecord` 接口增 `canonical_version`、`drift_status`。
2. `VersionBadge` 改由 `drift_status` 驱动（aligned/customized/outdated/unregistered）；`outdated` 用醒目橙色并写明「出厂已更新到 vX」。
3. 操作按钮按状态：
   - `outdated`：「**采用出厂新版**」→ 调 `getCanonical` 取出厂内容、预填编辑器、顶部横幅提示「这是出厂 vX，可直接保存采用，或改完再保存」。保存即采用并对齐分支点。同时保留「编辑」。
   - `customized`：「编辑」（预填维护版）+「重置为出厂」。
   - `aligned`：「编辑」。
4. `adminApi.ts` 增 `getCanonical(module)` → `GET /api/admin/prompts/:module/canonical`。

## 验证（TDD）

- **SyncCanonical**（`sync_test.go`，fakeStore 风格）：已存在的行——无论 `is_customized`、无论版本是否不一致——一律 noop 不被覆盖；缺失的行被补种。
- **drift_status 计算**：表驱动覆盖 aligned / customized / outdated / unregistered 四态。
- **保存逻辑**：存与出厂相同内容 → `is_customized=false` 且分支点=当前出厂；存不同内容 → `is_customized=true` 且分支点=当前出厂。
- **端到端复现成功判据**：本地构造「维护版自定义且分支点落后」的行 → 后台显示 `outdated` 橙标 → 点「采用出厂新版」保存 → 变 `aligned`/`customized`。

## 数据与上线

- **无表结构迁移**：复用 `version` / `canonical_hash` / `is_customized` / `content`。
- **线上即时修复顺带完成**：本设计上线后，线上合盘那行因 `canonical_hash != 当前出厂 Hash` 会显示 `outdated`，管理员点「采用出厂新版」即拿到 v3.1。

## 明确不改

- 合盘 `reading_id` 报告缓存（`GetLatestCompatibilityReport`）——与本问题无关。
- 运行时「DB 优先、代码兜底」的读取顺序——保持现状。

## 关联

与进行中的 OpenSpec change `prompt-template-versioning`（0/33）主题相关；落地时需决定本设计是并入该 change 还是新建 change。
