## Why

目前 Admin 后台大量使用了系统自带的 Emoji 作为图标。Emoji 虽能快速构建 UI，但具有“玩具感”，且在不同操作系统下的渲染表现不一（如 Windows 上尤为明显），这极大削弱了后台管理面板的专业性和高级质感。为提升产品的视觉标准，我们需要移除所有的 Emoji 标签，并引入现代化的矢量图标库。

## What Changes

- **前端依赖**：引入 `lucide-react` 图标库
- **移除 Emoji**：全面审查并移除 Admin 后台相关组件及页面中硬编码的 Emoji 字符
- **图标替换**：
  - 侧边栏导航：使用对应的 Lucide 图标替代
  - 数据面板卡片：使用精美、克制的线框图标
  - 核心状态标识（成功/失败等）：使用 CSS 几何圆点（Dot）或对应状态小图标
- **微调布局**：根据新图标的大小优化边距和对齐关系

## Capabilities

### New Capabilities
- `admin-ui-icons`: 引入并全量使用规范化、标准化的 SVG 图标体系来提升后台视觉体验

### Modified Capabilities
（此变更属于纯前端表现层优化，不涉及底层核心业务需求变更）

## Impact

- **前端依赖**：需要向 `package.json` 添加新依赖
- **影响范围**：修改 `AdminLayout.tsx`、`AdminDashboardPage.tsx`、`AdminLLMPage.tsx`、`AdminAILogsPage.tsx` 等数个管理员页面
- **风险极低**：纯样式渲染层面，不涉及数据获取与状态流转
