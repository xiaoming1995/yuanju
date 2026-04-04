## 1. 数据库解绑（解除历史覆盖）

- [x] 1.1 在 `backend/pkg/database/database.go` 中新增 Migration 片段，通过 `ALTER TABLE bazi_charts DROP CONSTRAINT IF EXISTS bazi_charts_chart_hash_user_id_key` 及等同索引命令剔除前期的复用防御隔离机制。
- [x] 1.2 在 `backend/internal/repository/repository.go` 中，修改 `CreateChart()` 以剥除 `ON CONFLICT DO UPDATE` 的逻辑段落，纯粹退回执行 `INSERT INTO ... RETURNING id` 并且无需进行查重验证干预。

## 2. API 资源模型重塑设计与重塑

- [x] 2.1 在 `backend/internal/repository/repository.go` 新增一个纯依靠主键查询记录的基础方法函数块 `GetChartByID(id string) (*model.BaziChart, error)`
- [x] 2.2 在 `backend/internal/handler/bazi_handler.go` 中改造 `Calculate` 的发牌表现，要求当有用户存在自动建立底层资源后，在构建 JSON 响应回传时候将建立的 `savedChart.ID` 装填并明牌返程 (`c.JSON` 中附加 `"chart_id": ...` 字典键)。
- [x] 2.3 修改 `backend/cmd/api/main.go`。将基于表单 Input 流水发函的报告路由 `POST /api/bazi/report` 转变为强制带有下钻层接的动态 RESTful 路由：即挂接至 `POST /api/bazi/report/:chart_id` 上。

## 3. 后端处理接口（Handler）迁移改造及风控

- [x] 3.1 对 `backend/internal/handler/bazi_handler.go` 中已经变身为 `/report/:chart_id` 形式的 `GenerateReport` 接口内部进行拆筋拔骨式的重构大动：首先剥离任何与提取及解析原有 `ReportInput` 等请求体动作的代码；
- [x] 3.2 调用刚刚新增的 `repository.GetChartByID(c.Param("chart_id"))` 来装取其对应的原始存储件并进行判空、且防范性校验其带有的 `chart.UserID` 指针是否恰巧等于当前通过了 `middleware.Auth` 并挂在 `c.Get("user_id")` 上的身份对象；如果越界或者错位应返回 403 此属主权校验。
- [x] 3.3 然后再用 `chart` 中提炼复刻出来的年／月／日／性别等数据就地组装进行 `bazi.Calculate()` 精算动作调用；将其结论作为入库或送 AI 组词的原料并与既有逻辑对齐走到底层直到取得分析报告返回。

## 4. 适配前端打通（消除暗雷 Bug）

- [x] 4.1 在 `frontend/src/lib/api.ts` 中，更新对应的接口层：使 `calculate` 的返回增加对于 `chart_id?: string` 的 TS 描述应对；其次彻底替换旧式的 `generateReport(input: CalculateInput)` 传参法则而迭代升级为仅需要依靠一串 UUID 字符请求对岸的方法 `generateReport(chartId: string)`； 并在底部配套变换发送路径。
- [x] 4.2 移步至 `frontend/src/pages/ResultPage.tsx`，将 `handleGenerateReport` 中的“如果无 `input` 即不可运作”此拦截铁门废弃拦截拆卸；而替换并引入判别条件：“必须确保在此页当前作用域范围内拥有一个已落地确诊的 `result.chart_id` 或从历史栏发过来的 params.id（我们统归获取存为一个最终决定调用的 `targetId` 发往服务端生算）。” 随后通过将该此 `id` 下发 `baziAPI.generateReport(targetId)` 呼叫完成贯通。
