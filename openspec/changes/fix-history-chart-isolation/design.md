## Context

`bazi_charts.chart_hash` 目前是单列 UNIQUE，由生辰数据（年月日时 + 性别）生成。`CreateChart` 使用 UPSERT：

```sql
ON CONFLICT (chart_hash) DO UPDATE SET user_id=EXCLUDED.user_id
```

这导致：不同用户相同生辰 → 后者"抢走"前者的记录；同一用户重复起盘 → `created_at` 不更新，历史显示旧日期。

`GetHistoryDetail` 只从 DB 读取 `BaziChart`（精简），未重新 `Calculate()`，导致 `ResultPage` 所需的 `year_gan_wuxing`、`hide_gan`、`shishen`、`di_shi`、`shen_sha` 等字段全部缺失。

## Goals / Non-Goals

**Goals:**
- 每个用户的命盘记录彼此隔离，相同生辰不会互相覆盖
- 同一用户重复对同一套生辰起盘时，复用已有记录（报告缓存有效）
- 历史详情页返回与正常起盘完全一致的完整 `result` 数据
- 增量 Migration，不丢失现有数据

**Non-Goals:**
- 跨用户公共命盘缓存（每个用户独立存储，AI 报告也独立）
- longitude / is_early_zishi 的持久化（本次不处理，重新 Calculate 时使用默认值）

## Decisions

### D1：改为 `(chart_hash, user_id)` 复合唯一约束

`chart_hash` 本身的语义保持不变（生辰指纹），只是约束从「全局唯一」变为「每用户唯一」。

```sql
-- 增量迁移（两步）
ALTER TABLE bazi_charts DROP CONSTRAINT IF EXISTS bazi_charts_chart_hash_key;
ALTER TABLE bazi_charts ADD CONSTRAINT bazi_charts_chart_hash_user_id_key
    UNIQUE (chart_hash, user_id);
```

存量数据若存在冲突（同一 chart_hash 多条 + 同 user_id），需先清理。由于目前数据量小（6 条），人工核查即可；Migration 代码加 `IF NOT EXISTS` 保持幂等。

### D2：UPSERT 冲突键和更新字段同步调整

```go
// 旧（危险）
ON CONFLICT (chart_hash) DO UPDATE SET user_id=EXCLUDED.user_id

// 新（安全）：用户重复起盘同一生辰时，只更新 AI 推断的用神信息
ON CONFLICT (chart_hash, user_id) DO UPDATE
  SET yongshen=EXCLUDED.yongshen, jishen=EXCLUDED.jishen
```

### D3：`GetChartByHash` 增加 `userID` 参数

当前无 user_id 过滤，跨用户可能读到他人记录。改为：

```go
func GetChartByHash(hash, userID string) (*model.BaziChart, error) {
    // SELECT ... WHERE chart_hash=$1 AND user_id=$2
}
```

调用方 `report_service.go` 需同步传入 userID。

### D4：`GetHistoryDetail` 补充 `bazi.Calculate()` 调用

Handler 拿到 BaziChart 后，用存储的 birth 数据重新计算，将完整 `result` 一并返回：

```go
result := bazi.Calculate(chart.BirthYear, chart.BirthMonth, chart.BirthDay,
    chart.BirthHour, chart.Gender, false, 0)

c.JSON(200, gin.H{
    "chart":  chart,
    "result": result,  // 新增
    "report": report,
})
```

前端 `ResultPage.tsx` 已有 `res.data.result || res.data.chart` 逻辑，自动生效，无需改动。

## Risks / Trade-offs

- **[longitude 精度损失]** 历史详情重算时 `longitude=0`，真太阳时不修正 → 极少数边界用户（时区边界 + 精确经度输入）可能出现 1 小时差异 → 暂接受，后续可加字段持久化
- **[存量数据冲突]** 若现有 6 条记录中有 `(chart_hash, user_id)` 重复 → 需人工清理后再 Migration → 当前数据量小，风险不高
- **[report_service 调用处改动]** `GetChartByHash` 签名变更，需全量检查调用方 → 只有 `report_service.go` 一处，可控
