## ADDED Requirements

### Requirement: 算法参数从数据库加载

系统 SHALL 在服务启动时从 `algo_config` 表加载算法数值参数到内存缓存。若表中无对应 key，SHALL 使用硬编码默认值作为 fallback，保证算法行为与历史一致。

#### Scenario: 服务启动时加载参数
- **WHEN** 服务启动，`algo_config` 表中存在 key `jixiong_jiHan_min`
- **THEN** `CalcDayunJixiong` 使用 DB 中的值作为极寒判定阈值

#### Scenario: DB 无配置时使用默认值
- **WHEN** `algo_config` 表为空或不含某 key
- **THEN** 算法使用硬编码默认值运行，行为与改动前一致

---

### Requirement: 管理员可查询当前算法参数

系统 SHALL 提供 `GET /api/admin/algo-config` 接口，返回所有算法参数的当前值（含来源：db / default）及描述说明。

#### Scenario: 获取参数列表
- **WHEN** 管理员调用 `GET /api/admin/algo-config`
- **THEN** 返回包含 key、value、source、description 的数组

---

### Requirement: 管理员可修改算法参数

系统 SHALL 提供 `PUT /api/admin/algo-config/:key` 接口，管理员可更新单个参数值。更新后 SHALL 立即写入 DB，但内存缓存不自动刷新，需手动触发 reload。

#### Scenario: 更新极寒判定阈值
- **WHEN** 管理员 PUT `{"value": "3"}` 到 `/api/admin/algo-config/jixiong_jiHan_min`
- **THEN** DB 中该 key 的 value 更新为 "3"，返回 200

#### Scenario: 写入非法值
- **WHEN** 管理员尝试写入非数字字符串到数值类型参数
- **THEN** 返回 400，说明期望格式

---

### Requirement: 管理员可热重载算法缓存

系统 SHALL 提供 `POST /api/admin/algo-config/reload` 接口，触发后重新从 DB 加载全部算法配置到内存，无需重启服务，新参数立即对后续计算生效。

#### Scenario: 热重载成功
- **WHEN** 管理员调用 reload 端点
- **THEN** 内存缓存更新为 DB 最新值，返回 200 及加载条数摘要
