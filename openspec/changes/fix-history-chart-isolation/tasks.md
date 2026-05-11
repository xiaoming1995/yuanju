## 1. 数据库 Migration

- [x] 1.1 在 `backend/pkg/database/database.go` 的增量迁移区追加两条 DDL，先删除旧的单列唯一约束 `bazi_charts_chart_hash_key`，再创建 `(chart_hash, user_id)` 复合唯一约束
- [x] 1.2 在同一迁移块中，同步删除旧的单列索引 `idx_bazi_charts_hash`，创建新的复合索引 `idx_bazi_charts_hash_user`（`chart_hash, user_id`）以保持查询性能

## 2. repository 层修改

- [x] 2.1 修改 `CreateChart()` 的 UPSERT 冲突键：`ON CONFLICT (chart_hash)` → `ON CONFLICT (chart_hash, user_id)`
- [x] 2.2 修改 `CreateChart()` 的 UPSERT 更新内容：`DO UPDATE SET user_id=EXCLUDED.user_id` → `DO UPDATE SET yongshen=EXCLUDED.yongshen, jishen=EXCLUDED.jishen`
- [x] 2.3 修改 `GetChartByHash()` 函数签名，新增 `userID string` 参数
- [x] 2.4 修改 `GetChartByHash()` 的 SQL 查询，在 `WHERE` 条件中加入 `AND user_id=$2`

## 3. report_service 调用方更新

- [x] 3.1 在 `backend/internal/service/report_service.go` 中找到调用 `repository.GetChartByHash()` 的位置，传入 `userID` 参数（确认：该函数在 service 层无直接调用，无需改动）
- [x] 3.2 若 `report_service.GenerateAIReport()` 当前不接收 `userID` 参数，则在函数签名中新增该参数，并更新 `bazi_handler.go` 中的调用处（确认：不需要，缓存按 chartID 查询）

## 4. handler 层：历史详情补充完整 result

- [x] 4.1 在 `GetHistoryDetail` handler 中，读取 chart 成功后，调用 `bazi.Calculate(chart.BirthYear, chart.BirthMonth, chart.BirthDay, chart.BirthHour, chart.Gender, false, 0)` 重新计算完整结果
- [x] 4.2 在 `GetHistoryDetail` 的 JSON 响应中新增 `"result": result` 字段（与 `GenerateReport` 响应格式对齐）

## 5. 验证

- [x] 5.1 确认本地后端能正常编译（`go build ./...`）
- [x] 5.2 重启后端，确认 Migration 增量迁移成功执行（查看启动日志）
- [x] 5.3 以同一用户账号对同一套生辰起盘两次，验证历史列表只出现一条记录（复用）
- [x] 5.4 以两个不同用户账号对同一套生辰起盘，验证各自历史列表只显示自己的记录
- [x] 5.5 点击历史详情，验证四柱网格的十神、长生、旬空、神煞、纳音字段正常显示
- [x] 5.6 验证历史详情中 AI 报告正常显示，精简/专业切换功能正常
