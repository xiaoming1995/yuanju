## 1. 安装依赖

- [x] 1.1 在 `frontend` 目录中运行 `npm install lucide-react` 安装图标库前端依赖

## 2. 改造 AdminLayout 整体布局

- [x] 2.1 引入 `lucide-react` 中的 `Hexagon` / `Settings`, `LayoutDashboard`, `Bot`, `Users`, `ListText` 等图标至 `src/components/AdminLayout.tsx`
- [x] 2.2 替换左侧导航栏以及顶部 Brand 名称处的所有硬编码 Emoji 符号。
- [x] 2.3 微调侧边栏 CSS 样式使其兼容 Flexbox + 图标垂直居中的效果（如果需要的话）。

## 3. 改造 Admin 页面内容中的 Emoji

- [x] 3.1 改造 `AdminDashboardPage.tsx`，将 `👥`, `☯`, `🤖`, `🛠️`, `🗑️`, `✅`, `❌` 替换为 `lucide-react` 组件并微调。
- [x] 3.2 改造 `AdminLLMPage.tsx`，替换页面 Title 的 `🤖` 等相关 Emoji。
- [x] 3.3 改造刚才我们编写的 `AdminAILogsPage.tsx`，将其中的标题 Emoji、按钮的对钩/叉号、卡片上的 Emoji 替换掉。

## 4. 验证效果

- [x] 4.1 在本地（或容器环境，取决于热重载）查看是否全部页面均不再出现由于系统不同而五颜六色的原生 Emoji 符号。
