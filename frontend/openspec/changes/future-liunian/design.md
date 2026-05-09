## Context

`GetPastYearSignals` 当前通过 `ln.Year > currentYear` 过滤掉未来年份，`GeneratePastEventsStream` 也用 `dy.StartYear > currentYear` 截断大运列表。前端 `PastEventsPage` 只渲染收到的年份数据，没有主动区分过往/未来，视觉上也无区分逻辑。

`DayunItem.LiuNian` 已包含完整的 10 年流年数据（含未来），算法引擎 `GetYearEventSignals` 对任意年份均可运行，数据层无需新增字段。

## Goals / Non-Goals

**Goals:**
- 移除后端过滤逻辑，让算法扫描全部流年（过往 + 未来）
- 前端在现有时间轴末尾展示未来年份，视觉与过往区分
- 旧缓存（仅含过往）失效后自动用新全量格式重新生成

**Non-Goals:**
- 不新增 API 端点或数据库表
- 不对未来年份使用不同语气（与过往叙述风格统一）
- 不限制未来覆盖年数（展示全部剩余大运）

## Decisions

### 1. 直接移除过滤条件，而非新增参数

**选择**：删除 `ln.Year > currentYear` 和 `dy.StartYear > currentYear` 的 break 逻辑，`GetAllYearSignals` 扫描全部流年。

**理由**：算法本身无状态，对未来年份计算结果同样有效。增加参数控制会引入不必要的调用方复杂度，且本功能明确要求覆盖全部剩余大运。

**备选**：新增 `includesFuture bool` 参数。拒绝——增加参数意味着两条代码路径需要各自维护。

### 2. 前端 isFuture 判断用当前年份常量

**选择**：前端 `parseAndSet` 时记录 `currentYear = new Date().getFullYear()`，渲染年份卡片时 `isFuture = y.year > currentYear`。

**理由**：前端无需后端告知哪些是未来年份，用客户端时间判断即可，逻辑简单且准确。

### 3. 未来卡片样式：虚线边框 + opacity 0.75 + 「未来」角标

**选择**：`border-style: dashed`，`opacity: 0.75`，右上角小角标「未来」。

**理由**：虚线在命理/占卜类 UI 中语义上代表"未定/待验证"，透明度降低视觉权重，角标明确标识性质，三者组合清晰区分过往（实线/不透明）与未来（虚线/半透明）。

## Risks / Trade-offs

- [AI 响应变长导致延迟和成本上升] → 全部流年（过往+未来）数据量约为原来 2-3 倍；可在 Prompt 中要求未来年份叙述精简（1-2 句）
- [旧缓存格式失效] → 清除所有 `ai_past_events` 记录；页面首次访问触发重新生成，耗时约 40-60 秒

## Migration Plan

1. 部署新后端（`GetAllYearSignals`）
2. 清空 `ai_past_events` 表：`DELETE FROM ai_past_events;`
3. 部署新前端（未来卡片样式）
4. 用户首次访问时自动重新生成全量数据并缓存
