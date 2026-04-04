## ADDED Requirements

### Requirement: Admin UI Standard Icons
后台所有的页面标题、卡片装饰、状态标识以及操作按钮 SHALL 必须使用无版权的、现代的单色矢量图标（如 `lucide-react` 图标）替代原生 Emoji，以提供专业严谨的后台界面视觉风格。

#### Scenario: 渲染系统导航菜单
- **WHEN** 用户在桌面端或者移动端展开 Admin 侧边栏
- **THEN** 系统导航菜单项呈现统一使用线框图（SVGs），并随 active 颜色进行变化联动，不存在 Emoji 的彩色突兀感。

#### Scenario: 渲染数据状态面板
- **WHEN** 渲染如“近7天总调用”、“清除缓存”或“成功标签”等元素
- **THEN** 不可再看到任何系统 emoji。操作采用 `size` 与字体对齐的标准矢量图，并在悬停产生颜色互动时拥有统一样式的动画效果。
