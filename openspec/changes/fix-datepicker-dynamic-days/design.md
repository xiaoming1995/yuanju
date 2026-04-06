## Context

起盘表单（`HomePage.tsx`）的「日」下拉框使用固定数组 `Array.from({ length: 31 })` 生成，完全不考虑月份或年份的实际天数。用户可以选择 2月30日、4月31日等不存在的日期并提交，后端收到非法日期后行为不确定（lunar-go 可能静默错位计算）。

**受影响路径**：`frontend/src/pages/HomePage.tsx` 第 93 行，仅前端逻辑。

## Goals / Non-Goals

**Goals:**
- 「日」下拉选项随年份和月份动态联动，永远只显示该月真实存在的日期
- 切换年份或月份后，若当前已选 `day` 超出新月实际天数，自动重置为 1
- 闰年 2 月正确显示 29 天，平年正确显示 28 天

**Non-Goals:**
- 不在后端增加日期合法性校验（后端 lunar-go 天文库本身有一定容错性，当前 scope 仅修前端）
- 不改变表单整体 UI 结构或交互流程
- 不涉及农历（阴历）的月份天数处理

## Decisions

### 决策：用 `new Date(year, month, 0).getDate()` 计算最大天数

这是 JavaScript 标准惯用法：

```
new Date(2024, 2, 0).getDate()  // = 29（2024年2月，闰年）
new Date(2023, 2, 0).getDate()  // = 28（2023年2月，平年）
new Date(2024, 4, 0).getDate()  // = 30（4月）
```

**原理**：`new Date(year, month, 0)` 表示 `month` 月的第 0 天，即上一月的最后一天，`.getDate()` 返回该天的日期数，等于上月天数。

**优点**：JS 内置，无需额外依赖，自动处理所有闰年边界。

### 决策：年/月 onChange 统一走独立的联动处理函数

不复用通用 `handleChange`，而是为年份和月份变更单独写处理逻辑，在更新年/月的同时检查并修复 `day`，避免 `day` 状态短暂越界。

## Risks / Trade-offs

- **无**：纯 UI 行为修复，不引入新依赖，不影响提交数据格式。
