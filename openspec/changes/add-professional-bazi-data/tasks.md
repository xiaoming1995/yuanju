## 1. Engine 层数据下发拓展
- [x] 1.1 在 `BaziResult` 结构体内增加 `Year/Month/Day/HourGanShiShen` 等字面量字段。
- [x] 1.2 在 `Calculate` 中调用 `lunar-go` 相应的方法 (`GetYearShiShenGan`、`GetYearShiShenZhi` 以及长生 `bz.GetYearDiShi()` 等) 填充字段并序列化返回。

## 2. 前端界面卡片重构
- [x] 2.1 修改 `frontend/src/pages/ResultPage.tsx`，更新接收 `BaziResult` 的 `interface`。
- [x] 2.2 在卡片 `map` 渲染中：在顶层天干处放置一小行 `span` 输出十神。
- [x] 2.3 在底层地支处放置十二长生（DiShi）和小字号的地支十神。在 CSS 中新增极细辅助字的对应类名保障美学平衡。
