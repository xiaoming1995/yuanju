## 1. 原生底层逻辑编写

- [x] 1.1 在 `backend/pkg/bazi/engine.go` 中，新增纯内部辅导函数 `inferNativeYongshen(dayGan string, stats *WuxingStats) (yongshen, jishen string)`。
- [x] 1.2 根据月令、干支五行关系以及生助比例编写 `if-else` 分离逻辑并返回汉字形式的五行串。

## 2. API 结构与流程挂载

- [x] 2.1 修改 `Calculate` 方法内部收尾处的装配阶段。在填充并统计好 `%` 之后，将算力传至 `inferNativeYongshen` 进行执行。
- [x] 2.2 给组装好的 `result.Yongshen` 和 `result.Jishen` 赋以默认推断出的初次结论。

## 3. 测试与效果打磨

- [x] 3.1 视察无登录权限的游客视角：确保立刻起盘后界面顶部即刻点亮喜用徽章。
- [x] 3.2 观察旧的数据库持久化策略逻辑不受影响（此时 `chart` 的 Yongshen 被 `Calculate` 自然带上，也会被存进 DB `bazi_charts` 原始记录）。
