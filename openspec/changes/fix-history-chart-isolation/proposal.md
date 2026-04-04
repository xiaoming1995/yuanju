## Why

历史记录模块存在两个联动 Bug，导致用户看不到今天生成的记录，且不同用户之间的命盘数据会互相污染：

1. **跨用户命盘污染**：`bazi_charts` 的 `chart_hash` 字段全局唯一，当两个用户分别对同一套生辰起盘时，UPSERT 会把 `user_id` 改写为后者，导致前者的历史记录消失，且后者看到的是前者生成的旧报告。

2. **「今天的记录看不见」**：即便是同一用户重复对相同生辰起盘，UPSERT 命中冲突后 `created_at` 不更新，历史列表仍显示旧日期，用户误以为新记录丢失。

3. **历史详情数据残缺**：`GetHistoryDetail` 只返回数据库存储的 `BaziChart`（精简结构），而结果页 `ResultPage` 需要完整的 `BaziResult`（含十神、藏干、纳音、长生、神煞等），导致从历史进入详情时大量字段显示异常。

## What Changes

- **数据库 Migration**：将 `chart_hash` UNIQUE 约束从单列改为 `(chart_hash, user_id)` 复合约束，实现每用户命盘隔离
- **`CreateChart()` UPSERT 更新**：冲突键改为 `(chart_hash, user_id)`，冲突时仅更新 `yongshen`/`jishen`（不再覆盖 `user_id`）
- **`GetChartByHash()` 增加 user_id 过滤**：查询时加入 user_id 条件，避免跨用户读取
- **`GetHistoryDetail` handler 补充 `bazi.Calculate()` 调用**：从存储的生辰数据重新计算，将完整 `result` 一并返回，修复详情页字段残缺问题

## Capabilities

### New Capabilities

（无新能力，均为 bug 修复）

### Modified Capabilities

- `bazi-history`：历史记录命盘数据 SHALL 按用户隔离存储，同一用户重复起盘相同生辰时复用已有记录，不同用户之间命盘互不可见
- `history-detail`：历史记录详情 SHALL 返回完整 `result`（含十神、藏干、纳音等精算字段），与正常起盘结果页数据一致

## Impact

- **后端**：`backend/pkg/database/database.go`（DDL Migration）、`backend/internal/repository/repository.go`（CreateChart + GetChartByHash）、`backend/internal/handler/bazi_handler.go`（GetHistoryDetail 补充 Calculate）
- **API**：`GET /api/bazi/history/:id` 响应新增 `result` 字段（向下兼容，前端已有处理逻辑）
- **无前端改动**：`ResultPage.tsx` 已有 `res.data.result || res.data.chart` 的降级逻辑
- **数据迁移注意**：需删除旧 UNIQUE 约束并创建新复合约束，存量数据若有冲突需先清理
