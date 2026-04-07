## ADDED Requirements

### Requirement: 使用 Lucide 统一线框图标渲染前端导航与标题

系统前端 SHALL 全面禁止使用 Emoji（如 ☯、📜、👤、🔮）作图标素材，改为使用 `lucide-react` 图标库中的专业矢量图标。

#### Scenario: 替换 Navbar 图标
- **WHEN** 用户查看系统 Navbar 导航栏
- **THEN** 系统显示 Lucide `Compass` 罗盘图作为 Logo 的引导图标，无任何表情符号。

#### Scenario: 替换 BottomNav 图标
- **WHEN** 移动端用户查看底部导航栏
- **THEN** "测算"显示 `Compass`，"历史"显示 `History`，"我的"显示 `User`。

#### Scenario: 替换内容卡片前缀
- **WHEN** 用户查看具体的命理分析卡片（如 ShareCard 标题、TiaohouCard 副标题）
- **THEN** 标题前缀从 🔮、📋 统一变更为微小的、发光的菱形 `Diamond` 或等效的仪表盘科技感极简单元素。

### Requirement: 五行徽章几何化体系

涉及五行属性渲染的组件（如 `YongshenBadge`）SHALL 停止使用具象 Emoji（✨🌲💧🔥🏔️），统一更换为着色的 Lucide `Hexagon` 图标，保持对原有五行基色的引用。

#### Scenario: 渲染喜用神徽章
- **WHEN** `YongshenBadge` 渲染用户的五行喜用神（例如"水"）
- **THEN** 界面上展示蓝色的 `Hexagon` 几何图标，不再出现具象的水滴 Emoji 💧。
