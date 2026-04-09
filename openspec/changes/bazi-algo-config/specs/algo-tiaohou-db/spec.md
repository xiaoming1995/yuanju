## ADDED Requirements

### Requirement: 调候用神规则从数据库加载

系统 SHALL 在服务启动时从 `algo_tiaohou` 表加载全部调候用神规则到内存。若表为空，SHALL 自动将现有硬编码字典（`tiaohou_dict.go` 中的 `tiaohouTable`）seed 至数据库，并从中加载，保证历史行为不变。

#### Scenario: 首次启动自动 seed
- **WHEN** 服务启动时 `algo_tiaohou` 表记录数为 0
- **THEN** 系统将 120 条硬编码默认规则写入表，并载入内存缓存

#### Scenario: 已有数据直接加载
- **WHEN** `algo_tiaohou` 表已有记录
- **THEN** 直接加载 DB 数据，不执行 seed，内存缓存反映最新 DB 内容

---

### Requirement: 调候规则 DB 覆盖硬编码

当 `algo_tiaohou` 表中存在某 `day_gan + month_zhi` 组合的记录时，系统 SHALL 优先使用 DB 值，忽略同 key 的硬编码值。

#### Scenario: DB 规则覆盖
- **WHEN** DB 中 `甲_子` 的 `xi_elements` 为 `丙,戊`（管理员修改过）
- **THEN** `calcTiaohou` 使用 `丙、戊` 而非硬编码的默认值

---

### Requirement: 管理员可查询调候规则列表

系统 SHALL 提供 `GET /api/admin/algo-tiaohou` 接口，返回全部 120 条（或更多）调候规则，支持按 `day_gan` 过滤。

#### Scenario: 查询全量规则
- **WHEN** GET `/api/admin/algo-tiaohou`
- **THEN** 返回 120 条记录数组，每条含 day_gan、month_zhi、xi_elements、text、source(db/default)

#### Scenario: 按日干过滤
- **WHEN** GET `/api/admin/algo-tiaohou?day_gan=甲`
- **THEN** 返回 12 条甲木对应的调候规则

---

### Requirement: 管理员可修改单条调候规则

系统 SHALL 提供 `PUT /api/admin/algo-tiaohou/:day_gan/:month_zhi` 接口，管理员可更新喜用天干列表与原文说明。更新后写入 DB，下次热重载后生效。

#### Scenario: 修改调候喜用天干
- **WHEN** 管理员 PUT `{"xi_elements": "丙,戊", "text": "..."}` 到 `/api/admin/algo-tiaohou/甲/子`
- **THEN** DB 中对应记录更新，返回 200

#### Scenario: 写入非法天干字符
- **WHEN** xi_elements 中包含非天干字符（如 "子"）
- **THEN** 返回 400，说明 xi_elements 仅接受合法天干字符（甲乙丙丁戊己庚辛壬癸）

---

### Requirement: 管理员可重置规则为默认值

系统 SHALL 提供 `DELETE /api/admin/algo-tiaohou/:day_gan/:month_zhi` 接口，将该条目从 DB 中删除，使其 fallback 到硬编码默认值。

#### Scenario: 删除自定义规则
- **WHEN** 管理员 DELETE `/api/admin/algo-tiaohou/甲/子`
- **THEN** DB 中该记录删除，下次热重载后该条目回归硬编码默认值，返回 200
