## Why

八字算命的「喜用神 / 忌神」是命局推断的核心产物，但目前仅以文字标签形式出现在 AI 报告中，缺乏视觉冲击力与个性化感知。用户拿不到一个"属于自己的"东西，复访意愿低，社交分享无抓手。现在 yongshen/jishen 字段已在后端稳定输出，是将其视觉化的最佳时机。

## What Changes

- 新增「命理色彩系统」：将五行（木/火/土/金/水）映射为固定的颜色、形状、方位语言，作为整个视觉化系统的基础规则集
- 新增「喜用神主题色标签」：在八字命盘结果页展示用户喜用神 / 忌神的五行属性徽章（含对应颜色、幸运色块、幸运数字、幸运方位）
- 新增「程序化命理头像」：基于用户喜用神的五行属性，使用 SVG + Canvas 前端程序化生成专属命理头像（无需外部 API，完全本地渲染），可供用户下载/保存

## Capabilities

### New Capabilities
- `wuxing-color-system`：五行 → 颜色/形状/方位/数字的静态映射规则集，作为所有视觉化功能的设计语言底座
- `yongshen-profile-badge`：在命盘结果页展示喜用神/忌神的个性化视觉徽章，包含属性标签、幸运色块、幸运数字、幸运方位
- `mingpan-avatar-generator`：基于五行喜用神程序化生成 SVG 头像，相同五行组合生成相同风格头像，用户可下载

### Modified Capabilities
（暂无现有能力规格需要修改）

## Impact

- **前端**：新增 `WuxingColorSystem.ts`（映射规则）、`YongshenBadge.tsx`（徽章组件）、`MingpanAvatar.tsx`（SVG 头像组件），集成到 `BaziResultPage` / `HomePage` 
- **后端**：无需修改，复用已有 `yongshen` / `jishen` 字段
- **数据库**：无需变更
- **外部依赖**：无（全部本地前端渲染，Zero API Cost）
