## 1. 全局样式支持
- [x] 1.1 在 `index.css` 或 `App.css` 添加 `.title-diamond-icon` 等支持标题前缀几何发光的 CSS 类

## 2. 导航组件替换
- [x] 2.1 修改 `Navbar.tsx`，将 `☯` 替换为 `lucide-react` 的 `Compass`，将样式居中对齐
- [x] 2.2 修改 `BottomNav.tsx`，将 `☯`、`📜`、`👤` 替换为 `Compass`、`History`、`User`，调整底栏对齐样式

## 3. 分析报表修饰替换
- [x] 3.1 修改 `YongshenBadge.tsx`，将五行标志 Emoji (✨🌲💧🔥🏔️) 替换为 Lucide `Hexagon`，并使用各自的五行颜色进行填充
- [x] 3.2 审查并替换 `ShareCard.tsx`、`TiaohouCard.tsx` 等存在 🔮、📋 标题修饰的前缀，改为 `<Diamond size={14} className="title-diamond-icon" />`
- [x] 3.3 审查 `BaziResult.tsx` 排盘详批等其他位置，将残余的零碎 emoji 清除干净
