## 1. 五行视觉映射规则集（基础底座）

- [x] 1.1 创建 `frontend/src/lib/wuxingColorSystem.ts`，定义 `WUXING_MAP` 常量，包含木/火/土/金/水各自的主色、深色、形状标识、方位、幸运数字数组、季节、质感关键词
- [x] 1.2 在同文件中实现 `parseWuxingList(str: string): WuxingKey[]` 工具函数，从字符串中提取有效五行字符（过滤非五行字符）
- [x] 1.3 导出 TypeScript 类型定义：`WuxingKey`（五行字符联合类型）、`WuxingProfile`（单条映射对象类型）

## 2. 喜用神属性徽章组件

- [x] 2.1 创建 `frontend/src/components/YongshenBadge.tsx`，接收 `yongshen` / `jishen` 字符串作为 props，解析并渲染五行属性标签行
- [x] 2.2 实现五行标签项：内含颜色圆点、五行汉字、Emoji，按照 `WUXING_MAP` 的主色设置背景色调
- [x] 2.3 合并喜用神各五行的幸运方位、幸运数字、季节展示在徽章卡片下半区
- [x] 2.4 实现点击标签展开/收起该五行释义文案的交互（inline展开）
- [x] 2.5 实现 `yongshen` 为空时的骨架屏占位状态，展示提示文案「生成 AI 报告后可解锁命元特质」
- [x] 2.6 为 `YongshenBadge` 编写对应的 CSS（使用 CSS Variables，匹配现有设计系统）

## 3. 程序化命理头像生成器

- [x] 3.1 创建 `frontend/src/components/MingpanAvatar.tsx`，接收 `yongshen`、`jishen`、`dayGan`（日主天干）作为 props
- [x] 3.2 实现 SVG 背景渐变层：喜用神第一五行主色 → 第二五行辅色（单五行时单色）
- [x] 3.3 实现五行纹样渲染函数（各五行各一个纹样生成器）：
  - 木：竖向竹节矩形组（`renderWoodPattern`）
  - 火：发散三角射线组（`renderFirePattern`）
  - 土：棋盘方格纹（`renderEarthPattern`）
  - 金：同心圆弧（`renderMetalPattern`）
  - 水：正弦波浪线（`renderWaterPattern`）
- [x] 3.4 实现 SVG 边框装饰层：忌神第一五行颜色描边
- [x] 3.5 实现中央日主天干汉字层：居中渲染大字，白色带轻微阴影
- [x] 3.6 实现 `yongshen` 为空时的锁定占位状态（显示锁定图标 + 文案）
- [x] 3.7 实现「下载头像」按钮，使用 `SVG → Canvas → canvas.toBlob()` 生成 PNG 下载；Canvas 不支持时降级为 SVG 下载

## 4. 集成到命盘结果页

- [x] 4.1 在 `BaziResultPage`（或 `HomePage` 的结果区块）中的五行分布图下方引入 `YongshenBadge` 组件
- [x] 4.2 在 `YongshenBadge` 下方引入 `MingpanAvatar` 组件预览区（含下载按钮）
- [x] 4.3 确保从 API 响应中正确传递 `yongshen`、`jishen`、`dayGan` 到两个新组件
- [x] 4.4 调整页面布局，确保新区块在移动端（375px）不破坏现有结构，字体控制适当

## 5. 视觉收尾与质量检查

- [x] 5.1 在桌面端（1280px）和移动端（375px）分别截图验证视觉效果
- [x] 5.2 测试喜用神/忌神各种边界情况：空值、单一五行、三个五行组合
- [x] 5.3 测试头像下载功能（Chrome / Safari）
- [x] 5.4 确认新增 CSS 未引入 magic number，全部使用 CSS Variables
