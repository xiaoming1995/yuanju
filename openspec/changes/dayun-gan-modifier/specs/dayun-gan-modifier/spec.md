## ADDED Requirements

### Requirement: 大运天干五行修正评级

系统 SHALL 在计算大运金不换评级时，同时考虑大运天干的五行属性，并根据该天干是否属于当前命盘的喜用或忌神方向，对地支评级进行升降调整，最终输出综合评级。

#### Scenario: 大运天干属于喜用方向五行
- **WHEN** 大运天干五行（如庚辛=金）属于命盘金不换规则的 GoodDirections 中任一方位的五行
- **THEN** 系统将地支评级提升一档（凶→平，平→吉，吉→大吉），并在 `gan_modifier` 字段返回 `"加成"`，`gan_desc` 说明加强原因

#### Scenario: 大运天干属于忌神方向五行
- **WHEN** 大运天干五行（如壬癸=水）属于命盘金不换规则的 BadDirections 中任一方位的五行
- **THEN** 系统将地支评级降低一档（大吉→吉，吉→平，平→凶，凶→大凶），并在 `gan_modifier` 字段返回 `"减损"`，`gan_desc` 说明减损原因

#### Scenario: 大运天干已达最高/最低评级时不再越界
- **WHEN** 地支原始评级为 `大吉` 且天干为加成，或地支原始评级为 `大凶` 且天干为减损
- **THEN** 综合评级维持不变，不超出评级上下界

### Requirement: 地支原始评级独立保留

系统 SHALL 在输出中同时保留地支独立评级，允许前端或数据分析层分别读取天干/地支各自的评价贡献。

#### Scenario: 返回数据包含 zhi_level 字段
- **WHEN** 任意命盘调用大运计算接口
- **THEN** 响应中 `dayun[].jin_bu_huan` 对象包含 `zhi_level` 字段，值为根据方位匹配得出的地支原始评级（不受天干影响）

### Requirement: 土干（戊/己）专属评价支持

系统 SHALL 为金不换规则的每个条目提供可选的土干专属评价字段 `earth_gan_eval`，当大运天干为戊/己时优先使用该字段评级；若字段为 `nil`，系统 SHALL 降级执行通关检测逻辑。

#### Scenario: earth_gan_eval 有数据时直接使用
- **WHEN** 命盘规则中 `earth_gan_eval` 字段不为 nil，且大运天干为戊或己
- **THEN** 系统直接使用 `earth_gan_eval.Level` 作为天干修正方向，给出 `gan_desc` 专项说明

#### Scenario: earth_gan_eval 为 nil 时启动通关检测
- **WHEN** 命盘规则中 `earth_gan_eval` 为 nil，且大运天干为戊或己
- **THEN** 系统检查 GoodDirections 中是否同时包含互克方位（如南方火+北方水），若是则 `gan_modifier` 返回 `"通关"`，评级提升；否则返回 `"中性"`，评级不变
