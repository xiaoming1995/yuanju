# 专业化文案与符号清理 Design Spec

## Goal

去掉用户端所有「AI」字眼、emoji 及 ✦ 等装饰性 Unicode 符号，使产品呈现为专业命理工具，而非 AI 产品。

## Scope

仅修改用户端（普通用户可见）的页面和组件。管理后台不做改动。

## 替换规则

### AI 字眼

| 原文 | 替换后 |
|------|--------|
| AI 命理解读 | 命理解读 |
| AI 合盘解读 | 合盘解读 |
| AI 智能解读 | 命理解读 |
| 生成 AI 命理解读 | 生成命理解读 |
| AI 正在深度推理中... 已思考 N 秒 | 正在深度推演中... 已思考 N 秒 |
| AI 正在生成本段大运总结…… | 正在生成本段大运总结…… |
| 大运 AI 总结 | 大运总结 |
| AI 测算中... | 测算中... |
| 获得完整 AI 解读报告 | 获得完整解读报告 |
| 获得 AI 智能解读报告 | 获得完整解读报告 |
| 大运总结由 AI 后台生成 | 大运总结后台生成 |
| AI 大运总结流中断：{err} | 大运总结生成中断：{err} |
| 结合 AI 生成完整解读 | 生成完整解读 |
| 尚未生成 AI 合盘解读 | 尚未生成合盘解读 |

### 人工智能描述（首页）

| 原文 | 替换后 |
|------|--------|
| 融合传统八字命理与人工智能，为你解读命盘中的天赋与机遇 | 融合传统八字命理与现代算法，为你解读命盘中的天赋与机遇 |

### Hero badge（首页）

| 原文 | 替换后 |
|------|--------|
| ✦ 八字命理 · AI 解读 ✦ | 八字命理 · 命理解读 |

### 特性卡片（首页）

| 原文 | 替换后 |
|------|--------|
| { icon: '✦', title: 'AI 智能解读', desc: '大模型结合命理知识，生成通俗易懂的个性报告' } | { icon: '', title: '命理解读', desc: '结合命理知识，生成通俗易懂的个性报告' } |
| { icon: '◉', title: '传统算法', ... } | { icon: '', title: '传统算法', ... } |
| { icon: '◈', title: '五行分析', ... } | { icon: '', title: '五行分析', ... } |

### 符号清理

| 原文 | 替换后 |
|------|--------|
| ✦ 立即起盘 | 立即起盘 |
| ✦ AI 命理解读（章节标题） | 命理解读 |
| ✦ 专属命理头像（章节标题） | 专属命理头像 |
| ✦ 命元特质 | 命元特质 |
| ✦ 精确交运时间：{date} ✦ | 精确交运时间：{date} |
| ✦ 缘 聚 命 理 ✦ | 缘 聚 命 理 |
| ✨ {year}年运势精批 | {year}年运势精批 |
| 🧠 AI 正在深度推理中 | 正在深度推演中 |
| ⚡（交脱日期前的图标） | 删除 |
| ⚠ {error} | {error} |
| report-summary-icon ✦ span | 删除该 span |

### 免责声明

| 原文 | 替换后 |
|------|--------|
| 本报告由 AI 辅助生成，内容仅供参考，不构成任何决策建议。 | 本报告内容仅供参考，不构成任何决策建议。 |
| 本推算基于八字命理算法与 AI 语言生成，仅供参考，不构成任何决策建议。 | 本推算内容仅供参考，不构成任何决策建议。 |

## 文件清单

### 页面文件

1. `frontend/src/pages/HomePage.tsx`
   - hero-badge 文字
   - hero-desc 文字
   - 提交按钮文字
   - guest-hint 文字
   - features 数组（icon、title、desc）

2. `frontend/src/pages/ResultPage.tsx`
   - 喜用/忌神加载占位文字
   - 章节标题（命理头像、命理解读）
   - report-summary-icon span
   - 加载状态文字（thinking）
   - 免责声明
   - 错误提示前缀 ⚠
   - 按钮文字「生成命理解读」
   - 未登录提示文字

3. `frontend/src/pages/PastEventsPage.tsx`
   - 状态栏文字（年份已就绪 · 大运总结...）
   - 底部统计文字
   - 大运总结加载占位
   - 大运总结流中断错误
   - 免责声明

4. `frontend/src/pages/CompatibilityResultPage.tsx`
   - 合盘解读章节标题
   - 空状态文字

5. `frontend/src/pages/CompatibilityPage.tsx`
   - 描述文字中的 AI 字眼

### 组件文件

6. `frontend/src/components/DayunTimeline.tsx`
   - ✦ 精确交运时间包裹符号
   - 大运分隔 ✦ span
   - ⚡ 交脱日期图标

7. `frontend/src/components/YongshenBadge.tsx`
   - ✦ 命元特质 标题

8. `frontend/src/components/ShareCard.tsx`
   - ✦ 缘 聚 命 理 ✦ 文字

9. `frontend/src/components/LiuYueDrawer.tsx`
   - ✨ {year}年运势精批 标题

## 不改动的内容

- 管理后台所有页面（AdminLLMPage、PromptSettings、AdminCelebritiesPage 等）
- 代码内部注释（如 `// AI 解读状态`）——仅改用户可见的 UI 文字
- 类型名称（如 `AIReport`）——仅改 UI 文字，不重构代码
- CSS class 名称
