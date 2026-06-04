# 全站 UX / 样式 / 布局优化 Implementation Plan

> **For agentic workers:** 这是执行蓝图，不是一次性实现清单。进入代码实现前，应先把目标批次转成 OpenSpec change，或明确只执行其中一个批次。

**Goal:** 基于已确认的「流程优先 + 设计系统收敛」方向，分批优化缘聚用户端和管理后台的 UX、样式一致性与布局合理性。

**Reference:** `docs/superpowers/specs/2026-06-01-ux-style-layout-audit-design.md`

**Mockup:** `.superpowers/mockups/yuanju-high-fidelity-ux-mockup.html`

**Tech stack:** React 18 + Vite + TypeScript + 纯 CSS Variables。禁止引入 UI 组件库。

**Verification baseline:**

```bash
cd frontend && npm run lint
cd frontend && npm run build
```

---

## Batch 0 · 基线与护栏

目标：先建立可回归的 UX / 样式基线，避免后续越改越散。

### Task 0.1: 记录当前问题基线

**Files:**
- Create: `docs/superpowers/audits/2026-06-01-frontend-ux-baseline.md`

**Steps:**

- [ ] 记录关键指标：`ResultPage.tsx` 行数、`ResultPage.css` 行数、inline style 数量、硬编码颜色数量、`alert/confirm` 数量。
- [ ] 按页面列出 P0 / P1 / P2 问题。
- [ ] 链接高保真静态稿。

### Task 0.2: 建立视觉验收截图目录

**Files:**
- Create: `docs/superpowers/audits/2026-06-01-screenshot-checklist.md`

**Steps:**

- [ ] 列出需要截图验收的页面：结果页、合盘入口、合盘结果、历史页、过往事件页、后台用户、后台 LLM、后台 AI 日志。
- [ ] 每页记录桌面宽度 1440px、移动端宽度 390px 的验收点。
- [ ] 说明截图只作为人工验收，不替代 lint/build。

---

## Batch 1 · 设计系统最小闭环

目标：先建共享组件和 token，不直接重排业务页面。

### Task 1.1: 扩展全局 CSS token

**Files:**
- Modify: `frontend/src/index.css`

**Steps:**

- [ ] 整理现有颜色变量，补齐 `surface`、`surface-muted`、`border`、`text-muted`、`success`、`warning`、`danger`。
- [ ] 定义字体层级变量：页面标题、区块标题、卡片标题、正文、说明文字、标签。
- [ ] 定义间距变量：页面、区块、卡片、表单行距。
- [ ] 定义统一圆角变量：按钮、输入框、卡片、弹窗。
- [ ] 保留当前深色金主题，不引入新主色。

### Task 1.2: 新增基础 UI primitives

**Files:**
- Create: `frontend/src/components/ui/PageShell.tsx`
- Create: `frontend/src/components/ui/PageShell.css`
- Create: `frontend/src/components/ui/SectionPanel.tsx`
- Create: `frontend/src/components/ui/SectionPanel.css`
- Create: `frontend/src/components/ui/Button.tsx`
- Create: `frontend/src/components/ui/Button.css`
- Create: `frontend/src/components/ui/StatusBadge.tsx`
- Create: `frontend/src/components/ui/StatusBadge.css`
- Create: `frontend/src/components/ui/EmptyState.tsx`
- Create: `frontend/src/components/ui/EmptyState.css`

**Steps:**

- [ ] 先实现无业务依赖的结构组件。
- [ ] 组件只使用 CSS variables，不写页面专属颜色。
- [ ] 保留 `className` 扩展入口，但不允许大段 inline style。
- [ ] 为按钮 variant 定义 `primary`、`secondary`、`ghost`、`danger`。

### Task 1.3: 新增反馈组件

**Files:**
- Create: `frontend/src/components/ui/ConfirmDialog.tsx`
- Create: `frontend/src/components/ui/ConfirmDialog.css`
- Create: `frontend/src/components/ui/Toast.tsx`
- Create: `frontend/src/components/ui/Toast.css`

**Steps:**

- [ ] `ConfirmDialog` 支持标题、描述、危险态、确认中状态。
- [ ] `Toast` 支持 success / error / info。
- [ ] 不引入第三方弹窗库。
- [ ] 后续批次逐页替换 `alert()` / `confirm()`。

**Verification:**

- [ ] `cd frontend && npm run lint`
- [ ] `cd frontend && npm run build`

---

## Batch 2 · 结果页首屏和行动结构

目标：先优化最核心用户页，减少寻找重点的成本。

### Task 2.1: 抽出结果页首屏摘要组件

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx`
- Modify: `frontend/src/pages/ResultPage.css`
- Create: `frontend/src/components/result/ResultHeroSummary.tsx`
- Create: `frontend/src/components/result/ResultHeroSummary.css`
- Create: `frontend/src/components/result/ResultActionBar.tsx`
- Create: `frontend/src/components/result/ResultActionBar.css`

**Steps:**

- [ ] 首屏展示四柱摘要、核心结论、主行动。
- [ ] 主行动固定为「生成 AI 解读」。
- [ ] 次行动为「查看过往事件」「导出 PDF」。
- [ ] 移动端底部固定主 CTA，避开 `BottomNav` 和 safe area。
- [ ] 不改八字计算数据结构。

### Task 2.2: 结果页分段导航收敛

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx`
- Modify: `frontend/src/pages/ResultPage.css`
- Create: `frontend/src/components/ui/SegmentedTabs.tsx`
- Create: `frontend/src/components/ui/SegmentedTabs.css`

**Steps:**

- [ ] 将结果页内容分为总览、命盘、用神、大运、AI 解读。
- [ ] 专业细节不再挤在首屏。
- [ ] 桌面使用页内锚点，移动端使用横向分段导航。
- [ ] 保证深链 hash 不破坏返回和导出。

### Task 2.3: 过往事件入口统一

**Files:**
- Modify: `frontend/src/pages/PastEventsPage.tsx`
- Modify: `frontend/src/pages/HistoryPage.tsx`
- Modify: `frontend/src/pages/ResultPage.tsx`

**Steps:**

- [ ] 统一「查看过往事件」入口文案。
- [ ] 统一历史记录卡片上的结果页 / 过往事件跳转。
- [ ] 空状态引导用户先生成命盘，不留空白列表。

**Verification:**

- [ ] 390px 宽度结果页无横向滚动。
- [ ] 首屏能看到核心结论和主行动。
- [ ] `cd frontend && npm run lint`
- [ ] `cd frontend && npm run build`

---

## Batch 3 · 合盘入口与合盘结果一致性

目标：让合盘从「填资料」变成「发起咨询」，并让结果阅读节奏和结果页一致。

### Task 3.1: 合盘入口三段式改造

**Files:**
- Modify: `frontend/src/pages/CompatibilityPage.tsx`
- Modify: `frontend/src/pages/CompatibilityPage.css`
- Reuse: `frontend/src/components/BirthProfileForm.tsx`

**Steps:**

- [ ] 第一段：咨询问题和关注主题。
- [ ] 第二段：关系阶段。
- [ ] 第三段：双方出生资料。
- [ ] 表单错误就地展示，不用 `alert()`。
- [ ] 提交按钮文案从技术动作改为用户动作。

### Task 3.2: 合盘结果视觉节奏对齐

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`
- Modify: `frontend/src/pages/CompatibilityResultPage.css`
- Reuse: `frontend/src/components/compatibility/*`

**Steps:**

- [ ] 保持「双方是谁 → 是否合 → 为什么 → 怎么做 → 专业依据」结构。
- [ ] 分数、verdict、下一步行动在首屏可见。
- [ ] 证据和专业细节允许折叠，但默认不隐藏关键判断。
- [ ] 分享卡和 PDF 不在本批次改动，除非文案不一致。

**Verification:**

- [ ] 合盘入口提交前用户能理解将生成什么。
- [ ] 合盘结果首屏能看到结论、分数和下一步。
- [ ] `cd frontend && npm run lint`
- [ ] `cd frontend && npm run build`

---

## Batch 4 · 管理后台一致性

目标：后台从页面级散装样式收敛成工具型界面。

### Task 4.1: Admin shell 和 Toolbar 统一

**Files:**
- Modify: `frontend/src/components/AdminLayout.tsx`
- Modify: `frontend/src/components/AdminLayout.css`
- Create: `frontend/src/components/admin/AdminPageHeader.tsx`
- Create: `frontend/src/components/admin/AdminToolbar.tsx`
- Create: `frontend/src/components/admin/AdminDataTable.tsx`
- Create: `frontend/src/components/admin/AdminDataTable.css`

**Steps:**

- [ ] 页面标题、说明、主操作按钮统一。
- [ ] 搜索、筛选、刷新、创建操作统一放入 Toolbar。
- [ ] 表格列间距、状态标签、行操作统一。

### Task 4.2: 替换浏览器原生弹窗

**Files:**
- Modify: `frontend/src/pages/admin/*.tsx`

**Steps:**

- [ ] 将 `window.confirm()` 替换为 `ConfirmDialog`。
- [ ] 将 `alert()` 替换为 `Toast` 或 inline error。
- [ ] 危险操作使用 danger button + 二次确认。
- [ ] 成功反馈不阻塞用户继续操作。

### Task 4.3: 管理后台状态标签统一

**Files:**
- Modify: `frontend/src/pages/admin/AdminUsersPage.tsx`
- Modify: `frontend/src/pages/admin/AdminLLMPage.tsx`
- Modify: `frontend/src/pages/admin/AdminAILogsPage.tsx`
- Modify: `frontend/src/pages/admin/AdminChartsPage.tsx`

**Steps:**

- [ ] active / inactive / success / failed / pending 使用统一 `StatusBadge`。
- [ ] 状态文案中文化，不混用英文状态和中文解释。
- [ ] 行操作不超过 3 个可见按钮，更多操作进入折叠菜单或次级样式。

**Verification:**

- [ ] `rg -n "alert\\(|confirm\\(" frontend/src/pages/admin` 无新增，目标逐步清零。
- [ ] 后台 1280px 宽度下核心表格字段不拥挤。
- [ ] `cd frontend && npm run lint`
- [ ] `cd frontend && npm run build`

---

## Batch 5 · 收尾页面和导出一致性

目标：让非核心页不再显得像另一套产品。

### Task 5.1: Auth / Profile 表单统一

**Files:**
- Modify: `frontend/src/pages/LoginPage.tsx`
- Modify: `frontend/src/pages/RegisterPage.tsx`
- Modify: `frontend/src/pages/ProfilePage.tsx`
- Modify: `frontend/src/pages/AuthPage.css`
- Modify: `frontend/src/pages/ProfilePage.css`
- Create: `frontend/src/components/ui/FormField.tsx`
- Create: `frontend/src/components/ui/FormField.css`

**Steps:**

- [ ] label、hint、error、submit loading 统一。
- [ ] 表单错误就地显示。
- [ ] 登录 / 注册 / 个人中心按钮层级一致。

### Task 5.2: HomePage 入口收敛

**Files:**
- Modify: `frontend/src/pages/HomePage.tsx`
- Modify: `frontend/src/pages/HomePage.css`

**Steps:**

- [ ] 首屏直接展示「开始排盘」「看历史」「合盘」三个入口。
- [ ] 减少纯装饰卡片，优先引导用户开始任务。
- [ ] 保留品牌氛围，但不牺牲入口清晰度。

### Task 5.3: 分享 / PDF 文案一致性检查

**Files:**
- Modify as needed: `frontend/src/components/ShareCard.tsx`
- Modify as needed: `frontend/src/components/CompatibilityShareCard.tsx`
- Modify as needed: `frontend/src/components/PrintLayout.tsx`
- Modify as needed: `frontend/src/components/CompatibilityPrintLayout.tsx`

**Steps:**

- [ ] 页面首屏结论和分享卡结论不互相矛盾。
- [ ] PDF 主标题、摘要、行动建议沿用页面层级。
- [ ] 不重做导出版式，只校正文案和结构一致性。

---

## Completion Checklist

- [ ] P0 页面已经完成：结果页首屏、后台反馈、基础组件。
- [ ] P1 页面已经完成：合盘入口、合盘结果、后台表格。
- [ ] inline style 数量有明确下降。
- [ ] `alert()` / `confirm()` 在后台页面中被清理或有剩余清单。
- [ ] `npm run lint` 通过。
- [ ] `npm run build` 通过。
- [ ] 390px、1440px 两个宽度完成截图验收。
- [ ] 文档更新为最终状态，标记哪些批次已完成。
