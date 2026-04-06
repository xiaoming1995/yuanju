## Why

目前在手机端（屏幕宽度 < 640px）时，顶部导航栏的 `.navbar-links` 被隐藏，导致手机端用户无法访问「历史记录」页面。引入底部标签导航栏（Bottom Tab Bar）不仅能立即解决历史记录入口丢失的问题，更契合原生 App 的移动端操作习惯，也为未来的功能扩展（如紫微斗数排盘等）提供更好的导航空间。

## What Changes

- **新增 `BottomNav` 组件**：仅在屏幕宽度 < 640px 时显示，悬浮于屏幕底部，包含「测算」和「历史」选项卡。
- **调整 `Navbar` 组件**：移除顶部领航栏在移动端的隐藏规则，直接隐藏顶部的 `.navbar-links`，并将空间让渡给新的底部导航栏。
- **全局布局调整**：为移动端 `body` 或全局容器底部增加安全内边距（如 `padding-bottom: 72px`），防止页面底部内容被新的悬浮栏遮挡。

## Capabilities

### New Capabilities
- `mobile-navigation`: 提供移动端专属的系统级底部悬浮导航。

### Modified Capabilities
- _无_

## Impact

- **前端组件**：新增 `frontend/src/components/BottomNav.tsx` 和对应的样式文件 `BottomNav.css`。
- **布局修改**：修改 `frontend/src/App.tsx` 的全局结构，引入 `<BottomNav />`。修改 `index.css` 提供底部边距。修改 `Navbar.css` 清除冗余。
- **外部依赖**：无需新依赖，纯 CSS 和 React 路由。
