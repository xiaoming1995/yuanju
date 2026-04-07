## Context

目前项目前端（React）组件中含有大量的硬编码表情符号（Emoji），用于起修饰、指示或区分不同元素（如五行）。项目已经引入了 `lucide-react` 组件库（Admin 后台已有使用）。为了统一“仪表盘、算法化、专业化”的产品调性，需要移除前端（尤其是用户 C 端）的所有 Emoji 渲染，使用 Lucide 中的几何或特定线框 SVG 替代。

## Goals / Non-Goals

**Goals:**
- 将 `Navbar`、`BottomNav` 上的导航 emoji 换成 Lucide 对应几何图标。
- 将涉及五行的组件（`YongshenBadge`）内的标志 Emoji 改为颜色的 `Hexagon` 图标填充。
- 将各分析内容卡片（`TiaohouCard`, `ShareCard`, `HistoryPage` 等）的前缀 emoji 去除，改配以统一样式的发光菱形 `Diamond` 加上适当的 CSS 变量修饰。
- 确保所有相关组件引用的依然是现有的五行颜色系统（`var(--color-wood)` 等）。

**Non-Goals:**
- 不涉及后台 Admin 的图标更改（后台目前已经是 Lucide 体系，状态良好）。
- 不更改原有的五行核心计算逻辑，仅修改视觉层渲染。
- 不引入其他外部图标库庞然大物。

## Decisions

### 1. 导航栏图标的映射配置
- **首页测算**：`Compass` 组件，寓意命理罗盘。
- **历史记录**：`History` 组件。
- **用户中心**：`User` 组件。

### 2. 五行徽章几何化
- 去除「火=🔥，水=💧」等映射，所有五行用神标记全部统一采用 `lucide-react` 中的 `Hexagon`（填充色继承五行基色），通过外加 `fill="currentColor"` 配合组件的 `color` 属性渲染出专业的科技质感。

### 3. 标题点缀统一标识
- 将原来分散的 🔮、📋、✨ 等全部废弃，设计一个通用的微型前缀：`<Diamond size={14} className="title-diamond-icon" />`
- 配合 CSS：`.title-diamond-icon { color: var(--color-primary); filter: drop-shadow(0 0 4px var(--color-primary-alpha)); }` 创造出仪表盘发光感。

## Risks / Trade-offs

- **引入 Lucide 带来的微小 Bundle Size 增加**：在 C 端引入几个 lucide icons 增加的大小非常微小，因为可以 Tree Shaking。
- **适配和对齐问题**：Emoji 和普通文本经常能 baseline 自动对齐，但 SVG 图标需要调整 flex 对齐方式（`display: flex; align-items: center`）。需要在 CSS 里仔细调整间距和对齐。
