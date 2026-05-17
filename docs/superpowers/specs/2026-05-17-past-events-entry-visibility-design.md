# 过往事件入口可见性 Spec

**日期：** 2026-05-17
**作者：** Claude + 用户
**状态：** 已批准，待实施

---

## 1. 背景与问题

`PastEventsPage`（`/bazi/:chartId/past-events`）是基于大运的年份级事件推算页，独立运行，**不依赖 AI 报告**——它只需要 `chartId` 与已登录用户即可加载（见 `frontend/src/pages/PastEventsPage.tsx:108` `baziAPI.fetchPastEventsYears`）。

但 `ResultPage` 中当前只有一个入口指向它：`ResultPage.tsx:1131-1135` 的 `report-action-bar` 内的按钮。该按钮被以下条件门控：

```tsx
{report && (
  ...
  {user && targetId && (
    <button onClick={() => navigate(`/bazi/${targetId}/past-events`)}>过往事件</button>
  )}
  ...
)}
```

后果：

- 用户起盘但未生成 AI 报告时，按钮**完全不渲染**；
- 即使报告生成完毕，按钮位于页面最底部 action bar 内，与"导出 PDF / 重新起盘"挤在一起，视觉权重低；
- 用户体验上"入口被藏起来了"。

`BottomNav`、`Navbar`、`HistoryPage` 也无任何 past-events 入口。

## 2. 目标

让"过往事件推算"入口在起盘成功后立刻可见，不再依赖 AI 报告先生成。

## 3. 非目标

- 不改 `PastEventsPage` 本身。
- 不动后端 / DDL / API。
- 不新增全局导航入口（`BottomNav` / `Navbar` 不在本次范围）。
- 不重排 `report-action-bar` 中其它按钮。

## 4. 方案

### 4.1 新入口位置

在 `ResultPage.tsx` 的 `<section className="dayun-section">` 内，紧跟 `<DayunTimeline />`（当前 `ResultPage.tsx:885-893`）插入一张引导卡。

**为什么放这里：**
- 大运是过往事件的载体（一个大运段展开 = 多个年份事件）。语义上紧贴 DayunTimeline 最自然。
- DayunTimeline 是起盘后立刻可见的核心组件，挂在它下面 = 第一屏可见。
- 不抢"命理解读"区块的主 CTA（"生成命理解读"按钮）视觉权重。

### 4.2 入口形态

一张轻量的横排引导卡（与下方 AI 解读 card 同体系，但更矮）：

```
┌──────────────────────────────────────────────┐
│ 🕰  过往事件推算                              │
│    展开每个大运段，看年份信号与白话批语    继续 →│
└──────────────────────────────────────────────┘
```

- **整卡可点击**（不止右侧）：`<button>` 或带 `role="button"` 的 `<div>`，点击触发 `navigate('/bazi/${targetId}/past-events')`。
- 使用现有 `card` 样式族，新增 className `past-events-entry`，定义 hover / focus / disabled 视觉态。
- 图标可用 lucide-react 已引入的图标（`History` 或 `Clock`）保持依赖不扩张。

### 4.3 渲染条件与状态

| 场景 | 渲染 | 状态 | 副标题 |
|------|------|------|--------|
| 已登录 + `targetId` 存在 | ✅ | 可点击 | "展开每个大运段，看年份信号与白话批语" |
| 未登录（guest）+ `result` 存在 | ✅ | 置灰 disabled | "登录后可查看" |
| 已登录但 `targetId` 还没拿到 | ❌ | — | — |
| `result` 不存在（起盘失败 / 还在加载） | ❌ | — | — |

未登录态点击不跳转（按钮 disabled），不做"点了再跳登录"——避免假阳性引导。

### 4.4 处理底部已有按钮

**删除** `ResultPage.tsx:1131-1135` `report-action-bar` 中的"过往事件"按钮。理由：

- 新入口位置更显眼且无门禁，底部按钮变成重复。
- `report-action-bar` 概念上应只放跟"报告本身"相关的动作（重新起盘 / 查看历史 / 导出 PDF），过往事件属于独立分析模块，混在那里语义不清。

`report && (...)` 包裹的其它按钮（重新起盘 / 历史 / 导出 PDF）保留不变。

## 5. 改动范围

| 文件 | 改动 |
|------|------|
| `frontend/src/pages/ResultPage.tsx` | dayun-section 内 DayunTimeline 之后插入新入口卡组件；删 `report-action-bar` 中过往事件按钮（约 5 行） |
| `frontend/src/pages/ResultPage.css` | 新增 `.past-events-entry`、`.past-events-entry:hover`、`.past-events-entry.is-disabled` 等样式 |
| `frontend/tests/past-events-entry-visibility.test.mjs`（新增） | 断言起盘成功后入口立即出现（不依赖报告）；未登录态时按钮 disabled |

零后端改动，零 DDL，零路由变更。

## 6. 验收标准

1. **匿名用户**：起盘 → ResultPage 加载完成 → DayunTimeline 下方出现"过往事件推算"入口卡，置灰，副标题为"登录后可查看"，点击无反应。
2. **登录用户**：起盘 → 入口卡可点击，副标题为"展开每个大运段，看年份信号与白话批语"，点击跳转至 `/bazi/:chartId/past-events`。
3. **登录用户报告生成后**：底部 `report-action-bar` 不再有"过往事件"按钮；保留"重新起盘 / 查看历史 / 导出 PDF"。
4. **键盘可达**：入口卡可通过 Tab 聚焦，Enter / Space 触发跳转（disabled 态不聚焦或聚焦但 Enter 无效）。
5. **响应式**：移动端宽度（≤480px）下卡片不溢出，右侧"继续 →"在窄屏可换行或隐藏文字保留箭头。

## 7. 风险与回退

- **风险低**：纯前端 UI 改动，无数据契约变更。
- **回退**：单文件 git revert 即可恢复原状。
- **测试覆盖**：新增的 mjs 测试保证入口可见性与门禁逻辑；不会引入新的服务端依赖。

## 8. 未来扩展（YAGNI 兜底，不在本次实施范围）

- 如果将来增加"未来流年推算"等同类模块，可在 DayunTimeline 下方堆叠多张引导卡，结构上预留好了。
- `BottomNav` 加 past-events 入口需要在 `chartId` 上下文外解决（如最近一次起盘的 chartId 缓存），不在本次范围。
