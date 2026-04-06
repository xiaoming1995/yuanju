## 1. 组件开发

- [x] 1.1 创建 `frontend/src/components/BottomNav.css`：实现底部固定布局、选项卡等分布局、并在宽度 >= 640px 时隐藏。
- [x] 1.2 创建 `frontend/src/components/BottomNav.tsx`：引入路由和 AuthContext，根据登录状态决定是否展示“历史”，动态匹配激活态。

## 2. 布局融合

- [x] 2.1 修改 `frontend/src/App.tsx`：引入 `BottomNav` 并将其与 `<Navbar />` 并列放在 `Router` 系统内。
- [x] 2.2 修改 `frontend/src/index.css`：增加全局或者针对 mobile 的 `@media (max-width: 640px)`，给外层容器 `padding-bottom: 70px;`（确保不受底部悬浮遮挡）。
- [x] 2.3 修改 `frontend/src/components/Navbar.css`：优化 `@media (max-width: 640px)` 规则，可能只需要维持隐藏 `.navbar-links` 即可，保持逻辑纯净。
