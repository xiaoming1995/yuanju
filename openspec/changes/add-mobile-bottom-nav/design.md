## Context

当前项目的 UI 采用响应式设计，但针对移动端（`max-width: 640px`）时，顶部 `<Navbar>` 组件仅通过 `display: none` 隐藏了 `.navbar-links`（包含“测算”、“历史”）部分。
这导致手机端用户登入后无法找到“历史记录”入口。为了提供更纯正的移动端体验，决定采用底部标签栏（Bottom Tab Bar）而非汉堡菜单。

## Goals / Non-Goals

**Goals:**
- 提供在移动端（< 640px）才会显示的底部导航组件（`<BottomNav />`）。
- 底部导航提供“首页(测算)”和“历史(如有权限)”切页。
- 确保底部导航固定底端时不遮挡原本页面底部的提交按钮等内容。

**Non-Goals:**
- 不重新设计桌面端的导航结构，桌面端继续沿用原本的 Navbar。
- 不引入第三方UI库框架如 Ant Design 移动端等，继续纯手写 CSS 变量。

## Decisions

### 1. 新组件 `BottomNav.tsx`
底部导航将在 `App.tsx` 中的 Router `<Navbar />` 下方添加，由于它是全局挂载并依赖于 `display` 和 Media Query 隐藏，我们可以直接把它放在全局。
**样式设计**：
- 采用 `position: fixed; bottom: 0; left: 0; right: 0;`。
- 高度设定为 `60px` 加上 env(safe-area-inset-bottom)。
- 提供激活态（匹配当前的 location.pathname），使用 CSS 高亮。

### 2. 平滑的安全区填补
为了避免底部由于引入 `BottomNav` 导致正文区域被遮挡。
- 在 `index.css` 的 `.page` 或者全局样式中添加媒体查询：移动端时，增加 `padding-bottom: calc(60px + env(safe-area-inset-bottom))`。

### 3. Navbar 的清理
将 `Navbar.css` 中的原本的 `@media` 内的 `.navbar-links` 的隐藏策略保留，确保移动端头部只显示 Logo 和 Logout 按钮，让出更多空间。

## Risks / Trade-offs

- **[风险] 状态和路由切换延迟** → 采用 `react-router-dom` 的 `<NavLink>` 解决激活状态显示问题。
- **[风险] 在非登录态下看到历史** → 跟 Navbar 类似，需要根据 `user` 的存在判断是否显示“历史”的 BottomNavItem。
