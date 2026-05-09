# Compatibility Result Display Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让合盘结果页先展示总评与双方命盘摘要，再展示评分、证据和 AI 解读。

**Architecture:** 复用现有 `participants[].chart_snapshot` 作为命盘摘要数据源，不改后端接口。前端在合盘结果页中新增双方命盘卡片与总评区块，并保留现有四维评分、证据和 AI 解读模块。

**Tech Stack:** React 19, TypeScript, Vite, CSS Variables

---

### Task 1: 补充前端展示类型

**Files:**
- Modify: `frontend/src/lib/api.ts`

- [ ] **Step 1: 定义合盘参与者命盘快照类型**

补充一个精简展示类型，至少覆盖：

```ts
export interface CompatibilityChartSnapshot {
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  gender: string
  day_gan: string
  year_gan: string
  year_zhi: string
  month_gan: string
  month_zhi: string
  day_zhi: string
  hour_gan: string
  hour_zhi: string
  wuxing?: { mu: number; huo: number; tu: number; jin: number; shui: number }
}
```

- [ ] **Step 2: 让 `CompatibilityParticipant.chart_snapshot` 使用该类型**

避免结果页继续依赖 `any`。

### Task 2: 重构合盘结果页信息层级

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

- [ ] **Step 1: 提取结果页展示辅助函数**

补充：

```ts
function formatBirthText(...)
function getParticipantSnapshot(...)
function getWuxingItems(...)
```

- [ ] **Step 2: 新增总评 Hero**

Hero 顶部展示：

```tsx
我 × 对方
契合等级文案
summary tags
一句短说明
```

- [ ] **Step 3: 新增双方命盘摘要双栏卡**

每张卡至少展示：

```tsx
身份名 + 性别
出生年月日时
四柱八字
日主
五行概览
```

- [ ] **Step 4: 保留并下移四维评分、关键证据、AI 解读**

维持现有功能，但阅读顺序调整为：

```text
Hero
双方命盘摘要
四维评分
关键证据
AI 解读
```

### Task 3: 样式与验证

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

- [ ] **Step 1: 收紧 badge 与卡片视觉层级**

让总评、命盘摘要和证据层级清晰，但不做复杂图表。

- [ ] **Step 2: 运行前端构建**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run build`  
Expected: PASS，允许大 chunk warning。

- [ ] **Step 3: 手工验证结果页**

检查：

```text
顶部先看到总评
中部能直接看到双方四柱八字
四维评分仍正常
关键证据和 AI 解读不受影响
移动端不会溢出
```
