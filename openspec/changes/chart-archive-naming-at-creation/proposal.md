# Chart Archive Naming At Creation

## Why
当前命盘档案命名（`display_name`）流程把"命名"塞在结果页底部，造成两类 UX 问题：

1. **信息架构问题**：`ResultPage.tsx:731-784` 的 `.chart-archive-tools` 卡片同时承担两个不相关的子任务——"给档案命名"和"用此命盘发起合盘"，且二者塞在同一 flex row。
2. **视觉层级问题**（2026-05-28 用户截图反馈）：
   - "档案称呼" 同时作为 `<h2 class="section-title">` 与字段 `<label>`，左列像两个并列标题
   - "用此命盘发起合盘" 是裸 `<span>`，夹在两个 `btn-ghost` 之间，无样式无可点击提示
   - 深色背景 + `btn-ghost` → "作为我 / 作为对方" 看不出是按钮

根因是把两个职责混在一个 section 里——"命名" 是档案创建的属性，"发起合盘" 是后续操作的入口。这两件事既然 HistoryPage 与 CompatibilityPage 已经各自提供了完整入口，结果页这个第三入口就是冗余的。

## What Changes
- **起盘表单新增可选"档案称呼"字段**（≤20 字符，UTF-8 rune count，规则与现有 `normalizeChartDisplayName` 一致）；用户在起盘时一次性填写
- **`POST /api/bazi/calculate` 入参增加 `display_name?: string`**；后端 `CalculateInput` 落库时一并写入 `BaziChart.DisplayName`
- **删除 `ResultPage.tsx` 中整块 `.chart-archive-tools` section**（lines 731-784）及其相关 state / handler / CSS 规则，包括：
  - "档案称呼" 编辑器（已挪到起盘前）
  - "用此命盘发起合盘 / 作为我 / 作为对方" 入口（已存在 HistoryPage + CompatibilityPage picker 两条入口）
- **保留** `PATCH /api/bazi/history/:id/display-name` 接口与历史/列表页的改名能力（用于事后修改）
- **保留** HistoryPage 的 "用此命盘合盘" 按钮与 CompatibilityPage 的 "从命盘档案选择" picker，行为不变

## Impact
- Affected specs:
  - **ADDED:** `bazi-chart-archive-naming`（new capability: 命盘档案命名时机）
- Affected code:
  - `frontend/src/components/BirthProfileForm.tsx`（新增 displayName input）
  - `frontend/src/pages/HomePage.tsx`（state + submit 拼接字段）
  - `frontend/src/lib/api.ts`（`CalculateInput.display_name?: string`）
  - `frontend/src/pages/ResultPage.tsx`（删 `.chart-archive-tools` section + 相关 state/handler）
  - `frontend/src/pages/ResultPage.css`（删 `.chart-archive-*` 规则与媒体查询分支）
  - `backend/internal/handler/bazi_handler.go`（`CalculateInput.DisplayName` + 复用 normalize + 落库填字段）
- DB migration: 无（`chart.display_name` 列早已存在）
- 兼容性:
  - 老命盘 `display_name` 仍可能为空，**不强制 backfill**
  - `PATCH /api/bazi/history/:id/display-name` 接口与 normalize 规则不变
  - 历史页与合盘页已有的"用此命盘合盘"/"从命盘档案选择"入口行为不变
