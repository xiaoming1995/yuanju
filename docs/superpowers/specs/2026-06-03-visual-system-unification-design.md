# 视觉系统统一首批设计

> Status: Approved for spec review
> Date: 2026-06-03
> Scope: 用户端 + 管理后台的基础视觉系统统一
> Related: `docs/superpowers/specs/2026-06-01-ux-style-layout-audit-design.md`

## 1. 背景

缘聚当前功能已经从 MVP 扩展到八字结果、历史、过往事件、合盘、个人中心、品牌导出和管理后台。前端经历多轮功能叠加后，页面级样式、按钮层级、状态反馈和弹窗行为开始分散。

已有 `2026-06-01` UX/style audit 已经覆盖用户流程和结果页首屏优化。本设计专门收敛首批“视觉系统统一”范围，目标是先建立可复用的 UI 基础设施，再逐页替换高频不一致点。

本批次不改变后端接口、不重排核心业务流程、不新增商业化能力。

## 2. 目标

首批优化目标：

- 统一颜色、字号、间距、圆角、边框、状态色等 CSS token。
- 让常用 UI 元素由共享组件承载，而不是在页面里重复写样式。
- 替换高频 `alert()` / `confirm()`，统一成功、失败、危险操作反馈。
- 降低用户端和管理后台的视觉割裂感。
- 为后续首页、合盘、个人中心和后台列表页改造提供稳定组件基础。

不追求一次性美化全站。首批成功标准是“新改动有统一基础可用，重点页面不再继续扩散散装样式”。

## 3. 非目标

本批次不做：

- 首页文案和起盘流程重排。
- 合盘入口三段式业务流程重排。
- 钱包、充值、点数、PDF 模板等新功能。
- 后端接口、数据库、权限逻辑改造。
- 全量重写所有 CSS。
- 引入 UI 组件库、TailwindCSS 或新设计框架。

## 4. 设计原则

### 4.1 Token 先行

页面不直接发明新颜色和尺寸。新增颜色、间距、字号、圆角、边框和状态色先进入 `frontend/src/index.css` 的语义 token，再由组件使用。

### 4.2 组件承载重复形态

按钮、区块、表单项、状态标签、空状态、确认弹窗和 Toast 属于重复形态，应沉到 `frontend/src/components/ui/`。页面只表达业务结构。

### 4.3 先替换高频痛点

优先处理会反复出现、影响用户感知或后台操作安全的元素：

- 浏览器原生 `alert()` / `confirm()`。
- 页面自定义按钮。
- 状态标签。
- 空、加载、错误状态。
- 大段 inline style。

### 4.4 保留现有品牌气质

继续使用深色金色方向，但金色只用于品牌、主行动、关键结论和选中态。普通边框、说明文字、后台状态不滥用金色。

## 5. 首批组件范围

### 5.1 CSS Token

扩展 `frontend/src/index.css`：

- Background: page, surface, surface-muted, overlay。
- Text: primary, secondary, muted, inverse。
- Border: default, subtle, accent。
- Action: primary, primary-hover, secondary, ghost。
- Status: success, warning, danger, info。
- Type scale: page title, section title, card title, body, caption, label。
- Spacing: page, section, card, form row, inline gap。
- Radius: button, input, card, dialog。

### 5.2 UI Primitives

首批共享组件：

- `Button`: primary, secondary, ghost, danger；支持 loading、disabled、icon。
- `PageShell`: 页面宽度、顶部间距、移动端安全区。
- `SectionPanel`: 区块标题、说明、内容容器。
- `FormField`: label、hint、error、必填状态。
- `SegmentedTabs`: 结果页、合盘页和后台分段切换。
- `StatusBadge`: success、warning、danger、info、neutral。
- `EmptyState`: 空数据、错误恢复、引导行动。
- `ConfirmDialog`: 危险操作确认、确认中状态。
- `Toast`: success、error、info 反馈。

组件保持业务中立，不写命理专属文案。

## 6. 首批页面落地范围

### 6.1 用户端

首批页面：

- 结果页：主行动按钮、分段导航、空/加载/错误状态、导出错误反馈。
- 历史页：空状态、删除确认、命盘状态标签、操作按钮。
- 个人中心：区块容器、功能卡状态、统计卡层级。

用户端首批不改起盘表单业务顺序，不重排合盘流程。

### 6.2 管理后台

首批页面：

- Admin LLM：Provider 状态、测试结果、删除确认、保存反馈。
- Admin 用户：禁用、解禁、重置密码、删除确认和反馈。
- Admin Prompt：保存、重置、漂移状态标签和错误反馈。

后台首批目标是工具一致性，不做大面积品牌装饰。

## 7. 数据流与状态

### 7.1 Toast

页面通过统一 Toast API 触发反馈：

- 成功保存：success。
- 请求失败：error。
- 长操作开始或说明性提示：info。

Toast 不替代页面内错误。表单字段错误仍就地展示，页面级加载失败使用 `EmptyState`。

### 7.2 ConfirmDialog

危险操作统一走确认弹窗：

- 删除命盘。
- 删除用户。
- 删除 Provider。
- 清除缓存。
- 重置 Prompt。

确认弹窗必须显示操作对象和后果。不可逆操作使用 danger 样式。

### 7.3 Loading

按钮级提交使用 `Button` loading 状态。页面级加载使用区块 skeleton 或 `EmptyState`，不使用浏览器原生弹窗。

## 8. 实施顺序

建议分四批：

1. Token 与 UI primitives：只建基础组件，不改业务页面。
2. 用户端替换：结果页、历史页、个人中心。
3. 管理后台替换：LLM、用户、Prompt。
4. 清理与验收：扫 `alert()` / `confirm()`、inline style、硬编码状态标签，补截图检查。

每批都应保持可单独 lint/build。

## 9. 验收标准

功能验收：

- 用户端和后台主要按钮样式统一。
- 后台危险操作不再使用浏览器原生确认框。
- 成功、失败反馈统一通过 Toast。
- 空、加载、错误状态不再留白或跳浏览器弹窗。
- 状态标签文案和颜色有统一映射。

工程验收：

- 不引入 UI 组件库。
- 新增/改造页面不继续引入大段 inline style。
- 新颜色先进入 token，不直接硬编码。
- `frontend/src/components/ui/` 组件无业务依赖。
- `cd frontend && npm run lint` 通过。
- `cd frontend && npm run build` 通过。

## 10. 风险与控制

风险：一次替换太多页面造成回归。

控制：按页面批次替换，每批只替换当前页面真正用到的组件。

风险：组件抽象过早导致页面难用。

控制：组件只覆盖高频形态；业务特有布局仍留在页面或领域组件。

风险：和当前未提交 UX 改动冲突。

控制：设计文档先独立提交。实现时先检查工作区状态，再决定接续已有改动还是开新变更。

