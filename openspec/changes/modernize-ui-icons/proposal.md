## Why

当前前端界面大量使用了 Emoji（如 ☯、📜、👤、🔮、五行符号 ✨🌲💧 等）来作为 UI 图标或章节前缀装饰。Emoji 存在多平台渲染不一致、色彩不可控、且带有一种随意的感观，这不符合打造“专业、精密、严肃”命理分析系统（偏向仪表盘科技感）的产品定位。因此，我们需要全局移除 Emoji，统一替换为极简、现代化的 SVG 矢量图标（Lucide 图标库）及几何点缀。

## What Changes

- 导航栏（Navbar / BottomNav）：将 ☯、📜、👤 替换为 `Compass`、`History`、`User` 纯线框图标
- 五行徽章与组件（YongshenBadge 等）：移除 ✨🌲💧🔥🏔️，改用极简的 `Hexagon`（六边形）配合五行主色调，不再具象化五行元素
- 卡片标题前缀（ShareCard / TiaohouCard 等）：移除 🔮、📋、✨ 等具体 Emoji，统一转换为纯粹的几何形态点缀（如发光的菱形 `Diamond` 或细垂直线指示器）
- 全局 CSS 优化：新增一些图标和前缀发光的 CSS 变量或类名，实现科技感

## Capabilities

### New Capabilities

- `modern-ui-icons`: 引入统一的 Lucide 线框式 SVG 图标策略与极简抽象的五行几何徽章设计，并移除前端组件中所有硬编码的 Emoji 字符。

### Modified Capabilities

- 无（仅视觉层替换，不改变任何业务逻辑与状态）

## Impact

- 涉及更改的所有使用 Emoji 的 React 组件，如 `Navbar.tsx`, `BottomNav.tsx`, `YongshenBadge.tsx`, `TiaohouCard.tsx`, `ShareCard.tsx`, `WuxingRadar.tsx` 等
- 需要稍微补充或修改组件对应的 CSS，以承载 SVG 图标的对齐与发光效果
