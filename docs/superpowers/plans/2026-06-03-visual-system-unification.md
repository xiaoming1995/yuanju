# 视觉系统统一首批 Implementation Plan

> For agentic workers: 这是执行蓝图，不是一次性全站重写清单。实现前必须先检查当前工作区状态，确认是否接续已有 UX 改动，不能覆盖用户未提交代码。

**Goal:** 按 `docs/superpowers/specs/2026-06-03-visual-system-unification-design.md`，统一缘聚用户端和管理后台的基础视觉系统，先建立 token 与 UI primitives，再替换重点页面中的高频不一致点。

**Scope:** React 18 + Vite + TypeScript + 纯 CSS Variables。禁止引入 UI 组件库和 TailwindCSS。

**Primary files likely touched:**

- `frontend/src/index.css`
- `frontend/src/components/ui/*`
- `frontend/src/pages/ResultPage.tsx`
- `frontend/src/pages/ResultPage.css`
- `frontend/src/pages/HistoryPage.tsx`
- `frontend/src/pages/HistoryPage.css`
- `frontend/src/pages/ProfilePage.tsx`
- `frontend/src/pages/ProfilePage.css`
- `frontend/src/pages/admin/AdminLLMPage.tsx`
- `frontend/src/pages/admin/AdminUsersPage.tsx`
- `frontend/src/pages/admin/PromptSettings.tsx`
- `frontend/src/lib/adminApi.ts`

**Verification baseline:**

```bash
cd frontend && npm run lint
cd frontend && npm run build
```

## Batch 0 · 工作区与现状确认

目标：先弄清当前已有 UX 改动，避免重复创建组件或覆盖未提交内容。

### Task 0.1: 检查工作区

Steps:

- [ ] 运行 `git status --short`。
- [ ] 如果已有 `frontend/src/components/ui/` 或 `frontend/src/components/result/` 未提交文件，先阅读它们。
- [ ] 记录哪些 UI primitives 已经存在，后续任务改为补齐/调整，不重复创建。
- [ ] 如果同一文件已有用户改动，基于现有内容最小化编辑，不回滚。

### Task 0.2: 采集替换基线

Steps:

- [ ] 统计 `frontend/src` 下 `alert(`、`confirm(`、`style={{`、硬编码状态文案。
- [ ] 记录本批要替换的页面和保留到后续批次的页面。
- [ ] 不因为统计结果顺手全站替换，只处理本计划范围内页面。

## Batch 1 · Token 与 UI primitives

目标：先建立可复用基础，不直接重排业务页面。

### Task 1.1: 扩展全局 CSS token

Files:

- Modify: `frontend/src/index.css`

Steps:

- [ ] 保留现有深色金色主题变量。
- [ ] 补齐语义 token：page/surface/surface-muted/overlay。
- [ ] 补齐文字 token：primary/secondary/muted/inverse。
- [ ] 补齐边框 token：default/subtle/accent。
- [ ] 补齐状态 token：success/warning/danger/info。
- [ ] 补齐字号、间距、圆角变量。
- [ ] 新 token 命名保持语义化，不使用页面名。

### Task 1.2: 补齐基础组件

Files:

- Create or modify: `frontend/src/components/ui/Button.tsx`
- Create or modify: `frontend/src/components/ui/Button.css`
- Create or modify: `frontend/src/components/ui/PageShell.tsx`
- Create or modify: `frontend/src/components/ui/PageShell.css`
- Create or modify: `frontend/src/components/ui/SectionPanel.tsx`
- Create or modify: `frontend/src/components/ui/SectionPanel.css`
- Create or modify: `frontend/src/components/ui/FormField.tsx`
- Create or modify: `frontend/src/components/ui/FormField.css`
- Create or modify: `frontend/src/components/ui/SegmentedTabs.tsx`
- Create or modify: `frontend/src/components/ui/SegmentedTabs.css`
- Create or modify: `frontend/src/components/ui/StatusBadge.tsx`
- Create or modify: `frontend/src/components/ui/StatusBadge.css`
- Create or modify: `frontend/src/components/ui/EmptyState.tsx`
- Create or modify: `frontend/src/components/ui/EmptyState.css`

Steps:

- [ ] 组件只依赖 React 和 CSS，不依赖业务类型。
- [ ] `Button` 支持 `variant`、`size`、`loading`、`disabled`、`icon`、`children`。
- [ ] `FormField` 支持 label、hint、error、required、children。
- [ ] `StatusBadge` 支持 neutral/success/warning/danger/info。
- [ ] `EmptyState` 支持 title、description、action。
- [ ] 保留 `className` 扩展入口。
- [ ] 避免组件内部写业务文案。

### Task 1.3: 补齐反馈组件

Files:

- Create or modify: `frontend/src/components/ui/ConfirmDialog.tsx`
- Create or modify: `frontend/src/components/ui/ConfirmDialog.css`
- Create or modify: `frontend/src/components/ui/Toast.tsx`
- Create or modify: `frontend/src/components/ui/Toast.css`

Steps:

- [ ] `ConfirmDialog` 支持 open、title、description、confirmLabel、cancelLabel、danger、loading。
- [ ] `Toast` 支持 success/error/info。
- [ ] 提供轻量 `useToast` 或 `ToastProvider`，不引入第三方库。
- [ ] 确认弹窗可键盘关闭，危险操作按钮有明确 danger 样式。

Verification:

- [ ] `cd frontend && npm run lint`
- [ ] `cd frontend && npm run build`

## Batch 2 · 用户端重点页面替换

目标：先处理用户最常访问的页面，统一按钮、状态和反馈，不改变业务流程。

### Task 2.1: 结果页替换

Files:

- Modify: `frontend/src/pages/ResultPage.tsx`
- Modify: `frontend/src/pages/ResultPage.css`

Steps:

- [ ] 将主要操作按钮替换为共享 `Button`。
- [ ] 将已有分段导航对齐到 `SegmentedTabs`。
- [ ] 将导出图片/PDF失败从 `alert()` 改为 Toast。
- [ ] 将加载/错误/未生成报告状态对齐到统一组件或统一样式。
- [ ] 不改报告生成接口、不改 AI 流式逻辑。

### Task 2.2: 历史页替换

Files:

- Modify: `frontend/src/pages/HistoryPage.tsx`
- Modify: `frontend/src/pages/HistoryPage.css`

Steps:

- [ ] 空历史使用 `EmptyState`。
- [ ] 删除命盘使用 `ConfirmDialog`。
- [ ] 删除成功/失败使用 Toast。
- [ ] 操作按钮使用共享 `Button`。
- [ ] 状态或标签使用 `StatusBadge`。

### Task 2.3: 个人中心替换

Files:

- Modify: `frontend/src/pages/ProfilePage.tsx`
- Modify: `frontend/src/pages/ProfilePage.css`

Steps:

- [ ] 页面外壳使用 `PageShell` 或对齐同等结构。
- [ ] 主要区块使用 `SectionPanel`。
- [ ] coming soon 功能状态使用 `StatusBadge`。
- [ ] 加载和错误状态使用统一样式，避免散装容器。
- [ ] 不新增钱包/充值/PDF 模板业务能力。

Verification:

- [ ] `cd frontend && npm run lint`
- [ ] `cd frontend && npm run build`

## Batch 3 · 管理后台重点页面替换

目标：后台操作反馈统一，危险操作不再使用浏览器原生弹窗。

### Task 3.1: Admin LLM 页面

Files:

- Modify: `frontend/src/pages/admin/AdminLLMPage.tsx`

Steps:

- [ ] Provider 激活/待机使用 `StatusBadge`。
- [ ] 删除 Provider 使用 `ConfirmDialog`。
- [ ] 保存、测试、删除结果使用 Toast。
- [ ] 表单按钮使用 `Button`。
- [ ] 不改 Provider API shape。

### Task 3.2: Admin 用户页面

Files:

- Modify: `frontend/src/pages/admin/AdminUsersPage.tsx`

Steps:

- [ ] 用户正常/禁用状态使用 `StatusBadge`。
- [ ] 禁用、解禁、删除使用 `ConfirmDialog`。
- [ ] 重置密码结果使用 Toast。
- [ ] 错误反馈不再使用 `alert()`。
- [ ] 不改后端用户管理接口。

### Task 3.3: Admin Prompt 页面

Files:

- Modify: `frontend/src/pages/admin/PromptSettings.tsx`

Steps:

- [ ] Prompt drift/customized/aligned 状态使用 `StatusBadge`。
- [ ] 保存和重置结果使用 Toast。
- [ ] 重置出厂版使用 `ConfirmDialog`。
- [ ] 保留现有 Prompt 编辑和 canonical 查看逻辑。

Verification:

- [ ] `cd frontend && npm run lint`
- [ ] `cd frontend && npm run build`

## Batch 4 · 清理与验收

目标：确认首批页面已经使用统一视觉系统，并记录未处理项。

### Task 4.1: 静态检查

Steps:

- [ ] 扫描本批页面是否仍有 `alert(` 或 `confirm(`。
- [ ] 扫描本批页面是否新增大段 `style={{ ... }}`。
- [ ] 扫描本批页面是否有新增硬编码状态色。
- [ ] 对保留项写明原因。

### Task 4.2: 视觉烟测

Steps:

- [ ] 打开结果页、历史页、个人中心，确认主要按钮、状态、空/错反馈一致。
- [ ] 打开 Admin LLM、Admin 用户、Admin Prompt，确认危险操作弹窗和 Toast 可用。
- [ ] 移动端宽度 390px 下确认用户端页面无横向滚动。
- [ ] 桌面端确认后台表单和按钮不拥挤。

### Task 4.3: 文档更新

Files:

- Modify or create audit note under `docs/superpowers/audits/`

Steps:

- [ ] 记录替换了哪些页面和组件。
- [ ] 记录剩余 `alert()` / `confirm()` 数量。
- [ ] 记录仍未处理的页面，作为下一批 UX 工作输入。

Final verification:

- [ ] `cd frontend && npm run lint`
- [ ] `cd frontend && npm run build`

## Risks

- 当前工作区已有未提交 UX 改动。实现前必须先阅读并接续，不得覆盖。
- 若 `frontend/src/components/ui/` 已有部分组件，以现有组件为准补齐能力。
- 不要把视觉系统统一扩大成全站重排；合盘流程、首页业务文案和商业化功能留给后续单独设计。

