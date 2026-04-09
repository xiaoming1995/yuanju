## Why

算法层（pkg/bazi）的关键判断参数（极寒极热阈值、身强判定线、调候用神规则）目前全部硬编码在 Go 源码中，命理专家无法在不部署代码的情况下修正规则。将这些规则迁移至数据库，由管理员后台统一管理，使算法行为可配置、可迭代。

## What Changes

- 新增 `algo_config` 数据库表，存储算法参数键值对（极寒/极热判定阈值、身强判定百分比等）
- 新增 `algo_tiaohou` 数据库表，存储 120 条调候用神规则（日干×月支 → 喜用天干+文字说明）
- 新增全局算法配置缓存层，服务启动时从 DB 加载规则到内存；DB 无配置时 fallback 至现有硬编码值
- 管理员后台新增「算法配置」页面，支持查看和修改算法参数与调候规则
- 新增管理员 API：`GET/PUT /api/admin/algo-config`、`GET/PUT /api/admin/algo-tiaohou/:key`

## Capabilities

### New Capabilities

- `algo-config`: 算法参数配置管理——管理员通过后台读写影响大运吉凶等计算的数值参数
- `algo-tiaohou-db`: 调候用神规则入库——将 120 条调候用神规则迁移至 DB，管理员可按日干+月支维度增删改查

### Modified Capabilities

（无现有 spec 层行为变更，算法接口不变，仅内部数据来源切换）

## Impact

- **后端**：`pkg/database/database.go`（新表 DDL）、`pkg/bazi/tiaohou_dict.go`（读缓存替换硬编码 map）、`pkg/bazi/dayun_jixiong.go`（阈值从缓存读取）、新增 `internal/service/algo_config_service.go`、新增 `internal/handler/algo_config_handler.go`、新增 `internal/repository/algo_config_repository.go`
- **前端**：管理员后台新增「算法配置」页面（两个子模块：参数配置、调候规则管理）
- **数据库**：新增两张表，无现有表结构变更
- **依赖**：无新增外部依赖
