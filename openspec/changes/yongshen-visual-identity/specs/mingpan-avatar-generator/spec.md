## ADDED Requirements

### Requirement: 程序化 SVG 命理头像生成
系统 SHALL 能够根据用户的喜用神（主色/辅色）和忌神（对比边框色）在浏览器端程序化生成一个 320×320 的 SVG 命理头像，无需任何外部 API 调用。

头像由以下层级构成：
1. 背景渐变层（喜用神第一五行主色 → 第二五行辅色，仅一个五行时单色）
2. 纹样层（根据五行形状规则生成几何图案）
3. 边框装饰层（忌神第一五行颜色，细描边）
4. 中央日主字符层（日柱天干汉字，白色/金色大字）

#### Scenario: 喜用神为木火时生成头像
- **WHEN** yongshen = "木火"，jishen = "金"，日主天干 = "甲"
- **THEN** 系统 SHALL 生成翠绿→暖橙的渐变背景、竹节纹样、冷银边框、中央"甲"字的 SVG 头像

#### Scenario: 喜用神仅单一五行
- **WHEN** yongshen = "水"（只有一个五行）
- **THEN** 系统 SHALL 生成单色深蓝背景 + 波浪纹样，无需渐变过渡

#### Scenario: yongshen 为空时不渲染头像组件
- **WHEN** yongshen 为空字符串
- **THEN** 系统 SHALL 不渲染头像预览区域，仅展示「生成 AI 报告后解锁命理头像」的锁定占位图标

### Requirement: 命理头像下载
系统 SHALL 提供「下载头像」按钮，点击后将 SVG 头像转换为 PNG 格式并触发浏览器下载，文件名为 `mingpan-avatar-<日主天干>.png`。

#### Scenario: 点击下载按钮
- **WHEN** 用户点击「下载头像」按钮且 SVG 已正确渲染
- **THEN** 浏览器 SHALL 触发文件下载，文件为 PNG 格式，分辨率不低于 320×320

#### Scenario: Canvas 渲染失败时降级
- **WHEN** 浏览器不支持 Canvas API 或 `toBlob` 方法
- **THEN** 系统 SHALL 降级为直接下载 SVG 文件，文件名为 `mingpan-avatar-<日主天干>.svg`
