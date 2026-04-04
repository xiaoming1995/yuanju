## Context

我们在上一阶段功能实现时为了速度和直观使用了大量的原生 Emoji（如侧边栏、卡片、操作按钮、状态提示）。为了追求更极致和专业的深色 Admin UI 设计感，我们决定引入目前流行的基于线框风格的开源图标库 `lucide-react`。

## Goals / Non-Goals

**Goals:**
- 将前台与后台割裂（本项目中目前明确优化的仅限 `/src/pages/admin/` 和 `AdminLayout` 内）
- 将以下常见业务语境的 Emoji 完全剔除并替换为相应 SVG Icon：
  - “数据概览”、“LLM 管理”、“用户列表”、“AI调用日志”、“清除缓存”等多处出现的情境。
  - “成功”、“失败”等表示系统状态类的图标。

**Non-Goals:**
- 不涉及大范围改变 DOM 层级结构或者对整站的颜色系统进行重构。
- 暂时不引入除了 `lucide-react` 之外的复杂 UI 组件库，依然坚持自行编写无侵入的 CSS。

## Decisions

### 1. 核心依赖
使用命令 `npm install lucide-react` 添加 SVG 矢量图标解决方案。所有引用的图标可按需导入，例如 `import { LayoutDashboard } from "lucide-react"`。

### 2. 映射表设计
为了保持视觉的一致性，我们会做如下统一的图标规范映射：
- 品牌 Logo：由 ⚙ 替换为 `Rocket` 或者 `LayoutTemplate` 图标结合文字
- 导航「数据概览」：由 📊 替换为 `LayoutDashboard`
- 导航「LLM 管理」：由 🤖 替换为 `Bot` 或 `Cpu`
- 导航「用户列表」：由 👥 替换为 `Users`
- 导航「AI 调用日志」：由 📋 替换为 `ListRender` 或者 `FileText`
- 警示及操作：由 🗑️ 替换为 `Trash2`，由 🛠️ 替换为 `Wrench`
- 状态展示：成功使用 `CheckCircle`，失败使用 `XCircle` 或 `AlertCircle`

### 3. CSS 微调
为了配合 SVG 的大小，SVG 图形一般设置 `size={18}` 或 `size={20}` 等，原先 `fontSize` 的设定将调整为通过 `display: flex` + `gap` 或者直接配合图标标签渲染。

## Risks / Trade-offs

- **引入第三方包负担**：`lucide-react` 大体上是支持 Tree Shaking 的，因此带来的产物体积增加极为微弱，可控且值得。
