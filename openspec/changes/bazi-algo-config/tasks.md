## 1. 数据库：新建两张表

- [x] 1.1 在 `pkg/database/database.go` 的 `Migrate()` 中添加 `algo_config` 表 DDL（key, value, description, updated_at）
- [x] 1.2 在 `Migrate()` 中添加 `algo_tiaohou` 表 DDL（day_gan, month_zhi, xi_elements, text, updated_at，复合主键）

## 2. 后端：算法配置缓存服务

- [x] 2.1 新建 `internal/repository/algo_config_repository.go`，实现 `algo_config` 表的 GetAll / Upsert 查询
- [x] 2.2 新建 `internal/repository/algo_tiaohou_repository.go`，实现 `algo_tiaohou` 表的 GetAll / Upsert / Delete 查询
- [x] 2.3 新建 `internal/service/algo_config_service.go`，实现内存缓存结构体（`sync.RWMutex`）及 `Load()` / `Reload()` 方法
- [x] 2.4 在 `algo_config_service.go` 中实现 `algo_tiaohou` 的首次启动 seed 逻辑（表为空时写入硬编码默认值）
- [x] 2.5 在 `pkg/bazi/` 中新建 `algo_config.go`，定义 `AlgoConfig` 结构体及 `SetAlgoConfig()` / `GetAlgoConfig()` 函数（全局变量，`sync.RWMutex` 保护）

## 3. 后端：算法层接入缓存

- [x] 3.1 修改 `pkg/bazi/dayun_jixiong.go`，将极寒/极热阈值、身强判定线从硬编码改为读 `GetAlgoConfig()`，fallback 到原有常量
- [x] 3.2 修改 `pkg/bazi/tiaohou_dict.go`，将 `tiaohouTable` 的读取改为优先从缓存取，fallback 到原有 Go map
- [x] 3.3 在 `cmd/api/main.go` 启动流程中，在路由注册前调用 `algo_config_service.Load()`，并将结果注入 `bazi.SetAlgoConfig()`

## 4. 后端：管理员 API

- [x] 4.1 新建 `internal/handler/algo_config_handler.go`，实现以下端点：
  - `GET /api/admin/algo-config`（返回全部参数及来源）
  - `PUT /api/admin/algo-config/:key`（更新单参数，含数值格式校验）
  - `POST /api/admin/algo-config/reload`（热重载缓存）
- [x] 4.2 新建 `internal/handler/algo_tiaohou_handler.go`，实现以下端点：
  - `GET /api/admin/algo-tiaohou`（支持 `?day_gan=` 过滤）
  - `PUT /api/admin/algo-tiaohou/:day_gan/:month_zhi`（更新单条，含天干字符校验）
  - `DELETE /api/admin/algo-tiaohou/:day_gan/:month_zhi`（删除自定义规则，恢复默认）
- [x] 4.3 在 `cmd/api/main.go` 中注册上述路由（挂载到 `adminAuth` 路由组）

## 5. 前端：管理员算法配置页面

- [x] 5.1 在 `src/pages/admin/` 新增 `AlgoConfigPage.tsx`，展示算法参数列表，支持行内编辑与保存
- [x] 5.2 在 `AlgoConfigPage.tsx` 中添加调候规则子模块，按日干分组展示 12 条规则，支持行内编辑喜用天干和原文
- [x] 5.3 添加「重载缓存」按钮，调用 reload 端点并提示成功/失败
- [x] 5.4 在 `adminApi.ts` 中添加 `algo-config` 和 `algo-tiaohou` 相关 API 方法
- [x] 5.5 在 `App.tsx` 中注册新页面路由，并在管理员侧边导航中添加入口

## 6. 验证

- [ ] 6.1 验证服务启动时 `algo_tiaohou` 表自动 seed 120 条记录
- [ ] 6.2 验证修改 `jixiong_jiHan_min` 参数 → reload → 新的 `CalcDayunJixiong` 使用新阈值
- [ ] 6.3 验证修改某条调候规则 → reload → `calcTiaohou` 返回更新后的喜用天干
- [ ] 6.4 验证 DELETE 调候规则 → reload → 该条目回归硬编码默认值
- [ ] 6.5 运行 `go test ./pkg/bazi/...` 确保全部通过
