# Prompt 出厂版/维护版两层模型 + 漂移可见 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 prompt 的「出厂版（代码 canonical，只读）」与「维护版（DB，运行时唯一源）」关系清晰可控——启动同步不再覆盖已存在的维护版，后台显示漂移状态并支持一键「采用出厂新版」。

**Architecture:** 运行时读取顺序不变（DB 优先、代码兜底）。改三处：(1) `SyncCanonical` 启动时对已存在行一律 noop，仅补种缺失行；(2) 新增纯函数 `prompt.DriftStatus` 在读取时比对 DB 行与代码 canonical 得出 3 态；(3) 后台保存把维护版「分支点」对齐到当前出厂版并据内容决定 `is_customized`。前端用 `drift_status` 驱动徽标与「采用出厂新版」流程。无数据库迁移。

**Tech Stack:** Go (Gin + database/sql + 标准库 crypto/sha256)，React + TypeScript（axios via `adminApi`）。

设计文档：`docs/superpowers/specs/2026-05-31-prompt-factory-maintained-split-design.md`

---

## 文件结构

| 文件 | 职责 | 动作 |
|---|---|---|
| `backend/pkg/prompt/canonical.go` | canonical 注册表 + 哈希助手 | 抽出 `HashContent` 助手 |
| `backend/pkg/prompt/drift.go` | 漂移状态纯函数 | 新建 |
| `backend/pkg/prompt/drift_test.go` | 漂移状态测试 | 新建 |
| `backend/pkg/prompt/sync.go` | 启动同步逻辑 | 改：已存在行一律 noop |
| `backend/pkg/prompt/sync_test.go` | 同步测试 | 改：升级断言改为不覆盖 |
| `backend/internal/repository/prompt_repository.go` | prompt 表读写 | 新增 `UpdateMaintained` |
| `backend/internal/handler/admin_prompt.go` | 后台 prompt API | 改 `GetPrompts`/`UpdatePrompt`；新增 `GetPromptCanonical` |
| `backend/cmd/api/main.go` | 路由注册 | 加 1 条 GET 路由 |
| `frontend/src/lib/adminApi.ts` | 前端 API 封装 | 加 `getCanonical` |
| `frontend/src/pages/admin/PromptSettings.tsx` | 后台 prompt 页面 | 徽标 + 采用流程 |

---

## Task 1: 抽出 `HashContent` 哈希助手

**Files:**
- Modify: `backend/pkg/prompt/canonical.go`（`Register` 内联哈希处，约 47-51 行）

漂移判定要对 DB 内容算 sha256，与 `Register` 里同一套逻辑。先抽成可复用的导出函数，避免重复。

- [ ] **Step 1: 新增 `HashContent` 并让 `Register` 复用它**

在 `canonical.go` 中，把 `Register` 里的两行哈希计算替换为调用新函数，并在文件中新增该函数（放在 `Register` 之后）：

```go
// HashContent 返回 content 的 sha256 hex 字符串，与 canonical Hash 同一算法。
func HashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
```

`Register` 内原本的：
```go
	sum := sha256.Sum256([]byte(def.Content))
	def.Hash = hex.EncodeToString(sum[:])
```
改为：
```go
	def.Hash = HashContent(def.Content)
```

- [ ] **Step 2: 编译验证**

Run: `cd backend && go build ./pkg/prompt/`
Expected: 成功无报错（`crypto/sha256`、`encoding/hex` 已 import，仍被 `HashContent` 使用）。

- [ ] **Step 3: 跑现有 prompt 包测试确认未回归**

Run: `cd backend && go test ./pkg/prompt/`
Expected: PASS（哈希值不变，所有现有测试照常通过）。

- [ ] **Step 4: Commit**

```bash
git add backend/pkg/prompt/canonical.go
git commit -m "refactor(prompt): 抽出 HashContent 助手供漂移判定复用"
```

---

## Task 2: `SyncCanonical` 对已存在行一律 noop

**Files:**
- Modify: `backend/pkg/prompt/sync.go`（`syncCanonicalWith`，约 51-82 行 + 头部注释 38-44 行）
- Modify: `backend/pkg/prompt/sync_test.go`（`TestSyncCanonical_UpgradesStaleAlignedRow`，约 118-141 行）

新模型：维护版神圣，启动同步只补种缺失行，**永不覆盖已存在行**（无论 `is_customized`、无论版本是否一致）。升级靠后台手动「采用」。

- [ ] **Step 1: 改写失败测试——已存在的旧版本行不应被覆盖**

把 `sync_test.go` 中的 `TestSyncCanonical_UpgradesStaleAlignedRow` 整个函数替换为：

```go
// TestSyncCanonical_DoesNotOverwriteExistingRow verifies the new model:
// any existing row (even a non-customized, version-mismatched one) is left
// untouched by startup sync. Upgrades happen only via admin "adopt factory".
func TestSyncCanonical_DoesNotOverwriteExistingRow(t *testing.T) {
	store := newFakeStore()
	store.rows["compatibility"] = &model.AIPrompt{
		Module:       "compatibility",
		Version:      "v2-old",
		IsCustomized: false,
		Content:      "old content",
	}
	if err := syncCanonicalWith(store); err != nil {
		t.Fatal(err)
	}
	if len(store.updates) != 0 {
		t.Errorf("existing row must never be overwritten, got updates=%v", store.updates)
	}
	if len(store.inserts) != 0 {
		t.Errorf("existing row must not be re-inserted, got inserts=%v", store.inserts)
	}
	if store.rows["compatibility"].Version != "v2-old" {
		t.Errorf("existing row version must stay v2-old, got %s", store.rows["compatibility"].Version)
	}
	if store.rows["compatibility"].Content != "old content" {
		t.Errorf("existing row content must stay unchanged, got %s", store.rows["compatibility"].Content)
	}
}
```

- [ ] **Step 2: 跑测试确认它失败**

Run: `cd backend && go test ./pkg/prompt/ -run TestSyncCanonical_DoesNotOverwriteExistingRow -v`
Expected: FAIL —— 当前 `sync.go` 仍会对版本不匹配的非自定义行调用 `UpdateCanonicalContent`，故 `store.updates` 非空。

- [ ] **Step 3: 改写 `sync.go` 的决策逻辑**

把 `syncCanonicalWith` 中的 `switch` 块（约 59-79 行）替换为：

```go
		switch {
		case row == nil:
			if err := store.InsertCanonical(module, def.Version, def.Content, def.Hash, def.Description); err != nil {
				log.Printf("[prompt-sync] module=%s action=skip reason=insert_error err=%v", module, err)
				continue
			}
			log.Printf("[prompt-sync] module=%s action=insert version=%s hash=%s", module, def.Version, shortHash(def.Hash))

		default:
			// 维护版神圣：已存在的行永不被出厂版自动覆盖。
			// 升级靠后台手动「采用出厂新版」；漂移由 DriftStatus 在读取时暴露。
			log.Printf("[prompt-sync] module=%s action=noop reason=row_exists version=%s", module, row.Version)
		}
```

并把函数上方的决策表注释（约 38-44 行）替换为：

```go
// Decision per module:
//
//	DB 无该模块 → InsertCanonical（补种初始值）
//	DB 已存在   → noop（维护版神圣，永不覆盖；升级靠后台手动「采用」）
//
// Errors are logged per-module; a failing module never blocks startup.
```

> 说明：`UpdateCanonicalContent`（接口/realStore/repo）保留不动——它现在不再被 sync 调用，但 fakeStore 的 `store.updates` 仍是测试断言「没发生覆盖」的探针。

- [ ] **Step 4: 跑全部 prompt 包测试**

Run: `cd backend && go test ./pkg/prompt/ -v`
Expected: PASS。`TestSyncCanonical_DoesNotOverwriteExistingRow` 通过；`TestSyncCanonical_SkipsCustomizedRow`、`TestSyncCanonical_NoOpOnAlignedRow`、`TestSyncCanonical_InsertsMissingModule`、两个 DBError 测试照常通过。

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/prompt/sync.go backend/pkg/prompt/sync_test.go
git commit -m "feat(prompt): 启动同步不再覆盖已存在维护版，仅补种缺失行"
```

---

## Task 3: `prompt.DriftStatus` 漂移状态纯函数

**Files:**
- Create: `backend/pkg/prompt/drift.go`
- Create: `backend/pkg/prompt/drift_test.go`

读取时比对 DB 行与代码 canonical，得出 4 态。纯函数，零 DB 依赖，表驱动测试。

判定（设计文档）：
- `canonical_hash != 出厂.Hash` → `outdated`（分支点落后于当前出厂版）
- 否则 `content` 哈希 == 出厂.Hash → `aligned`
- 否则 → `customized`
- 模块未注册 → `unregistered`

- [ ] **Step 1: 写失败测试**

`backend/pkg/prompt/drift_test.go`：

```go
package prompt

import "testing"

func TestDriftStatus(t *testing.T) {
	def := MustGet("compatibility") // 已注册模块，取当前出厂版

	cases := []struct {
		name          string
		module        string
		content       string
		canonicalHash string
		want          string
	}{
		{
			name:          "aligned: 内容与分支点都等于当前出厂版",
			module:        "compatibility",
			content:       def.Content,
			canonicalHash: def.Hash,
			want:          DriftAligned,
		},
		{
			name:          "customized: 基于当前出厂版但内容被改",
			module:        "compatibility",
			content:       "管理员改过的文案",
			canonicalHash: def.Hash,
			want:          DriftCustomized,
		},
		{
			name:          "outdated: 分支点落后于当前出厂版",
			module:        "compatibility",
			content:       "随便什么旧内容",
			canonicalHash: "old-hash-from-v2",
			want:          DriftOutdated,
		},
		{
			name:          "outdated: 分支点为空(历史遗留行)也算落后",
			module:        "compatibility",
			content:       def.Content,
			canonicalHash: "",
			want:          DriftOutdated,
		},
		{
			name:          "unregistered: 代码无此 canonical",
			module:        "no_such_module",
			content:       "x",
			canonicalHash: "y",
			want:          DriftUnregistered,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DriftStatus(tc.module, tc.content, tc.canonicalHash)
			if got != tc.want {
				t.Errorf("DriftStatus(%q,...) = %q, want %q", tc.module, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd backend && go test ./pkg/prompt/ -run TestDriftStatus -v`
Expected: FAIL —— `DriftStatus`/`DriftAligned` 等未定义，编译不过。

- [ ] **Step 3: 实现 `drift.go`**

```go
package prompt

// Drift 状态：维护版（DB 行）相对当前出厂版（代码 canonical）的关系。
const (
	DriftAligned      = "aligned"      // 维护版内容 == 当前出厂版
	DriftCustomized   = "customized"   // 基于当前出厂版，但内容被管理员改过
	DriftOutdated     = "outdated"     // 分支点落后于当前出厂版（出厂已更新）
	DriftUnregistered = "unregistered" // 代码侧无此 canonical（历史遗留）
)

// DriftStatus 比对 DB 行（dbContent + dbCanonicalHash 分支点）与当前出厂版，
// 得出漂移状态。纯函数，不触 DB。
func DriftStatus(module, dbContent, dbCanonicalHash string) string {
	def, ok := Lookup(module)
	if !ok {
		return DriftUnregistered
	}
	if dbCanonicalHash != def.Hash {
		return DriftOutdated
	}
	if HashContent(dbContent) == def.Hash {
		return DriftAligned
	}
	return DriftCustomized
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `cd backend && go test ./pkg/prompt/ -run TestDriftStatus -v`
Expected: PASS（全部 5 个子用例）。

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/prompt/drift.go backend/pkg/prompt/drift_test.go
git commit -m "feat(prompt): 新增 DriftStatus 纯函数判定维护版漂移状态"
```

---

## Task 4: repo `UpdateMaintained`

**Files:**
- Modify: `backend/internal/repository/prompt_repository.go`（在 `UpdatePrompt` 之后新增）

后台保存维护版时，一次性写 content / version / canonical_hash / is_customized——把「分支点」对齐到当前出厂版。

- [ ] **Step 1: 新增 `UpdateMaintained`**

在 `prompt_repository.go` 中 `UpdatePrompt` 函数之后插入：

```go
// UpdateMaintained 保存后台维护版：写 content，并把「分支点」(version/canonical_hash)
// 对齐到当前出厂版，据内容是否偏离出厂版设置 is_customized。
// 行通常已存在；ON CONFLICT 兜底首次场景。
func UpdateMaintained(module, content, version, hash string, isCustomized bool) error {
	_, err := database.DB.Exec(
		`INSERT INTO ai_prompts (module, content, description, version, is_customized, canonical_hash)
		 VALUES ($1, $2, '', $3, $4, $5)
		 ON CONFLICT (module) DO UPDATE
		   SET content = EXCLUDED.content,
		       version = EXCLUDED.version,
		       is_customized = EXCLUDED.is_customized,
		       canonical_hash = EXCLUDED.canonical_hash,
		       updated_at = NOW()`,
		module, content, version, isCustomized, hash,
	)
	return err
}
```

- [ ] **Step 2: 编译验证**

Run: `cd backend && go build ./internal/repository/`
Expected: 成功无报错。

- [ ] **Step 3: Commit**

```bash
git add backend/internal/repository/prompt_repository.go
git commit -m "feat(prompt): 新增 UpdateMaintained 保存维护版并对齐分支点"
```

---

## Task 5: `GetPrompts` 返回 `drift_status` + `canonical_version`

**Files:**
- Modify: `backend/internal/handler/admin_prompt.go`（`GetPrompts`，约 13-20 行）

列表每项附带计算字段，供前端徽标使用。用 handler 层 DTO，不污染 `model.AIPrompt`。

- [ ] **Step 1: 改写 `GetPrompts`**

把 `admin_prompt.go` 的 `GetPrompts` 函数替换为：

```go
// promptWithDrift 在 model.AIPrompt 之上附加读取时计算的漂移信息。
type promptWithDrift struct {
	model.AIPrompt
	CanonicalVersion string `json:"canonical_version"`
	DriftStatus      string `json:"drift_status"`
}

// GetPrompts 获取所有配置的 Prompts（含相对出厂版的漂移状态）
func GetPrompts(c *gin.Context) {
	prompts, err := repository.GetAllPrompts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取 Prompt 失败"})
		return
	}
	out := make([]promptWithDrift, 0, len(prompts))
	for _, p := range prompts {
		item := promptWithDrift{AIPrompt: p}
		if def, ok := prompt.Lookup(p.Module); ok {
			item.CanonicalVersion = def.Version
		}
		item.DriftStatus = prompt.DriftStatus(p.Module, p.Content, p.CanonicalHash)
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}
```

确认文件头 import 已含 `"yuanju/internal/model"`；若无则添加（`prompt` 与 `repository` 已 import）。

- [ ] **Step 2: 编译验证**

Run: `cd backend && go build ./internal/handler/`
Expected: 成功无报错。

- [ ] **Step 3: 跑 handler 包测试确认未回归**

Run: `cd backend && go test ./internal/handler/`
Expected: PASS（现有 `admin_compat_test.go` 等不受影响）。

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/admin_prompt.go
git commit -m "feat(prompt): GetPrompts 返回 drift_status 与 canonical_version"
```

---

## Task 6: `UpdatePrompt` 保存改写（堵 footgun + 对齐分支点）

**Files:**
- Modify: `backend/internal/handler/admin_prompt.go`（`UpdatePrompt`，约 22-49 行）

保存即「对照当前出厂版确认」：分支点对齐当前出厂版；内容与出厂版一字不差 → `is_customized=false`（堵住「原样保存即上锁」），否则 true。

- [ ] **Step 1: 改写 `UpdatePrompt`**

把 `UpdatePrompt` 函数替换为：

```go
// UpdatePrompt 更新特定的 Prompt 模板（保存维护版）。
// 分支点对齐当前出厂版；据内容是否偏离出厂版决定 is_customized。
func UpdatePrompt(c *gin.Context) {
	module := c.Param("module")
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	def, ok := prompt.Lookup(module)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "未知模块: " + module})
		return
	}

	isCustomized := prompt.HashContent(req.Content) != def.Hash
	if err := repository.UpdateMaintained(module, req.Content, def.Version, def.Hash, isCustomized); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新 Prompt 失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}
```

> 注：原先的 `repository.UpdatePrompt` + `repository.SetCustomized` 两步调用被 `UpdateMaintained` 取代。`log` import 若在本文件仅此处使用，移除以免编译错误（`ResetPromptToCanonical` 未用 log）。编译报错则按提示删 `"log"`。

- [ ] **Step 2: 编译验证**

Run: `cd backend && go build ./internal/handler/`
Expected: 成功。若报 `"log" imported and not used`，从 `admin_prompt.go` import 块删除 `"log"`。

- [ ] **Step 3: 跑 handler 测试**

Run: `cd backend && go test ./internal/handler/`
Expected: PASS。

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/admin_prompt.go
git commit -m "fix(prompt): 保存对齐分支点并堵住原样保存即上锁的 footgun"
```

---

## Task 7: `GetPromptCanonical` 端点 + 路由

**Files:**
- Modify: `backend/internal/handler/admin_prompt.go`（文件末尾新增）
- Modify: `backend/cmd/api/main.go`（约 239-241 行附近）

供前端「采用出厂新版」把出厂内容载入编辑器。

- [ ] **Step 1: 新增 handler**

在 `admin_prompt.go` 末尾追加：

```go
// GetPromptCanonical 返回指定模块的出厂版（代码 canonical）内容，供后台「采用出厂新版」载入编辑器。
func GetPromptCanonical(c *gin.Context) {
	module := c.Param("module")
	def, ok := prompt.Lookup(module)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown module: " + module})
		return
	}
	c.JSON(http.StatusOK, gin.H{"version": def.Version, "content": def.Content})
}
```

- [ ] **Step 2: 注册路由**

在 `main.go` 的 prompt 路由块（`adminAuth.POST("/prompts/:module/reset", ...)` 那行之后）加：

```go
				adminAuth.GET("/prompts/:module/canonical", handler.GetPromptCanonical)
```

- [ ] **Step 3: 编译验证**

Run: `cd backend && go build ./...`
Expected: 成功无报错。

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/admin_prompt.go backend/cmd/api/main.go
git commit -m "feat(prompt): 新增 GET /admin/prompts/:module/canonical 取出厂版内容"
```

---

## Task 8: 前端 `adminApi.getCanonical`

**Files:**
- Modify: `frontend/src/lib/adminApi.ts`（`adminPromptsAPI`，约 114-120 行）

- [ ] **Step 1: 加 `getCanonical`**

把 `adminPromptsAPI` 替换为：

```ts
export const adminPromptsAPI = {
  list: () => adminApi.get('/api/admin/prompts'),
  update: (module: string, data: { content: string }) =>
    adminApi.put(`/api/admin/prompts/${module}`, data),
  resetToCanonical: (module: string) =>
    adminApi.post(`/api/admin/prompts/${module}/reset`),
  getCanonical: (module: string) =>
    adminApi.get(`/api/admin/prompts/${module}/canonical`),
}
```

- [ ] **Step 2: 类型检查**

Run: `cd frontend && npx tsc --noEmit`
Expected: 无新增报错。

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/adminApi.ts
git commit -m "feat(prompt): 前端 adminApi 增加 getCanonical"
```

---

## Task 9: 前端 PromptSettings 漂移徽标 + 采用出厂新版流程

**Files:**
- Modify: `frontend/src/pages/admin/PromptSettings.tsx`

徽标改由 `drift_status` 驱动；`outdated` 时提供「采用出厂新版」（载入出厂内容到编辑器 + 横幅），`customized` 时保留「重置为出厂」。

- [ ] **Step 1: 扩展 `PromptRecord` 接口**（约 8-17 行）

在接口中加两个字段：

```ts
interface PromptRecord {
  id: string
  module: string
  content: string
  description: string
  version: string
  is_customized: boolean
  canonical_hash: string
  updated_at: string
  canonical_version: string
  drift_status: 'aligned' | 'customized' | 'outdated' | 'unregistered'
}
```

- [ ] **Step 2: 改写 `VersionBadge` 由 `drift_status` 驱动**（约 19-47 行）

替换整个 `VersionBadge` 函数：

```tsx
function VersionBadge({ record }: { record: PromptRecord }) {
  let color: string
  let bg: string
  let text: string

  switch (record.drift_status) {
    case 'outdated':
      color = '#f97316'
      bg = 'rgba(249,115,22,0.14)'
      text = `出厂已更新到 ${record.canonical_version}（你基于 ${record.version}）`
      break
    case 'customized':
      color = '#f59e0b'
      bg = 'rgba(245,158,11,0.12)'
      text = `已自定义（基于出厂 ${record.canonical_version}）`
      break
    case 'aligned':
      color = '#10b981'
      bg = 'rgba(16,185,129,0.12)'
      text = `已是出厂版 ${record.canonical_version}`
      break
    default:
      color = '#888'
      bg = 'rgba(136,136,136,0.15)'
      text = '历史遗留'
  }

  return (
    <span style={{
      fontSize: 11, color, background: bg,
      padding: '2px 8px', borderRadius: 4,
      marginLeft: 8, fontWeight: 500,
    }}>
      {text}
    </span>
  )
}
```

- [ ] **Step 3: 加「采用出厂新版」编辑横幅状态 + 处理函数**

在组件内 state 区（约 70-71 行 `resetLoading` 之后）加：

```tsx
  const [editBanner, setEditBanner] = useState<string | null>(null)
```

把 `handleEdit`（约 87-90 行）改为进入编辑时清掉横幅：

```tsx
  const handleEdit = (p: PromptRecord) => {
    setEditingModule(p.module)
    setEditContent(p.content)
    setEditBanner(null)
  }
```

在 `handleEdit` 之后新增「采用出厂新版」处理函数：

```tsx
  const handleAdoptFactory = async (module: string) => {
    try {
      const { data } = await adminPromptsAPI.getCanonical(module)
      setEditingModule(module)
      setEditContent(data.content)
      setEditBanner(`这是出厂最新版 ${data.version}，可直接保存采用，或在此基础上改完再保存。`)
    } catch (e: unknown) {
      alert(errorMessage(e, '载入出厂版失败'))
    }
  }
```

在 `handleSave`（约 92-105 行）成功分支里，保存后清横幅——把 `setEditingModule(null)` 一行后补一行 `setEditBanner(null)`。

- [ ] **Step 4: 按 `drift_status` 渲染操作按钮**（约 176-210 行 header 右侧按钮区）

把按钮区（`<div style={{ display: 'flex', alignItems: 'center', gap: 0 }}>` 内部）替换为：

```tsx
          <div style={{ display: 'flex', alignItems: 'center', gap: 0 }}>
            {p && !isEditing && p.drift_status === 'outdated' && (
              <button
                onClick={() => handleAdoptFactory(p.module)}
                style={{
                  marginLeft: 8, flexShrink: 0,
                  padding: '6px 14px', borderRadius: 6, fontSize: 13,
                  background: '#f97316', border: 'none',
                  color: '#000', fontWeight: 500, cursor: 'pointer',
                }}
              >
                采用出厂新版
              </button>
            )}
            {p && !isEditing && p.drift_status === 'customized' && (
              <button
                onClick={() => handleReset(p.module)}
                style={{
                  marginLeft: 8, flexShrink: 0,
                  padding: '6px 14px', borderRadius: 6, fontSize: 13,
                  background: 'transparent',
                  border: '1px solid #f59e0b',
                  color: '#f59e0b',
                  cursor: 'pointer',
                }}
              >
                重置为出厂
              </button>
            )}
            {p && !isEditing && (
              <button
                onClick={() => handleEdit(p)}
                style={{
                  marginLeft: 8, flexShrink: 0,
                  padding: '6px 14px', borderRadius: 6, fontSize: 13,
                  background: 'transparent',
                  border: '1px solid var(--color-border)',
                  color: 'var(--color-text-secondary)',
                  cursor: 'pointer',
                }}
              >
                编辑
              </button>
            )}
            {!p && !loading && (
              <span style={{ color: '#ff6b6b', fontSize: 12, flexShrink: 0 }}>⚠ 未初始化（请重启后端）</span>
            )}
          </div>
```

- [ ] **Step 5: 编辑区顶部显示横幅**（编辑分支内，约 218 行 `isEditing` 块开头）

在编辑分支最外层 `<div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>` 的**第一个子元素位置**插入：

```tsx
              {editBanner && (
                <div style={{
                  fontSize: 12, color: '#f97316',
                  background: 'rgba(249,115,22,0.08)',
                  padding: '10px 14px', borderRadius: 8,
                  border: '1px solid rgba(249,115,22,0.25)',
                }}>
                  📦 {editBanner}
                </div>
              )}
```

- [ ] **Step 6: 类型检查 + 构建**

Run: `cd frontend && npx tsc --noEmit && npm run build`
Expected: 无类型错误，构建成功。

- [ ] **Step 7: 手动验证（关键成功判据）**

启动后端 + 前端，登录后台 → AI 指令设定 → 批断指令 → 婚恋合盘解读：

1. **复现线上场景**：在本地 DB 把 compatibility 行的 `canonical_hash` 改成一个旧值（模拟分支点落后）：
   ```sql
   UPDATE ai_prompts SET canonical_hash = 'stale-hash', is_customized = true WHERE module = 'compatibility';
   ```
   刷新后台 → 该模块应显示橙色徽标「出厂已更新到 vX（你基于 …）」+「采用出厂新版」按钮。
2. 点「采用出厂新版」→ 编辑器载入出厂内容 + 顶部橙色横幅 → 点「保存更改」。
3. 刷新 → 徽标变绿「已是出厂版 vX」（`aligned`）。
4. 再点「编辑」改一个字保存 → 徽标变琥珀「已自定义（基于出厂 vX）」（`customized`），出现「重置为出厂」。
5. 点「重置为出厂」→ 回到绿色 `aligned`。

Expected: 上述状态流转全部符合。

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/admin/PromptSettings.tsx
git commit -m "feat(prompt): 后台按 drift_status 显示漂移徽标与采用出厂新版流程"
```

---

## 收尾验证

- [ ] **后端全量测试**

Run: `cd backend && go test ./...`
Expected: PASS（重点 `./pkg/prompt/`、`./internal/handler/`）。

- [ ] **前端构建**

Run: `cd frontend && npm run build`
Expected: 成功。

- [ ] **端到端**：按 Task 9 Step 7 完整走一遍状态流转。

---

## 自查记录（写作时已核对）

- **Spec 覆盖**：出厂/维护两层（Task 2 sync noop）、漂移 3 态可见（Task 3+5+9）、采用出厂新版不丢改动（Task 7+9，载入出厂内容后可在编辑器再改）、堵 footgun（Task 6）、无表迁移（全程复用现有列）、线上即时修复顺带完成（Task 9 验证即复现线上场景）——均有对应任务。
- **类型一致**：`DriftStatus`/`DriftAligned`/`DriftCustomized`/`DriftOutdated`/`DriftUnregistered`（Task 3）、`HashContent`（Task 1）、`UpdateMaintained` 五参签名（Task 4→6 一致）、前端 `drift_status` 联合类型（Task 9 Step 1）与徽标/按钮分支取值一致。
- **明确不改**：`reading_id` 报告缓存、运行时「DB 优先代码兜底」读取顺序、`ResetToCanonical`/reset 端点、`UpdateCanonicalContent`（保留作测试探针）。
