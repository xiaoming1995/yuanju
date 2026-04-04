## ADDED Requirements

### Requirement: 名人数据库与后台管理平台
支持管理员在 Admin 后台手动添加和编辑经过专业筛选的名人八字资料，供 AI 报告引用。

#### Scenario: 创建新的名人记录
- **WHEN** 管理员在 "名人库" 页面填写完整的名人姓名及命理特征（如“七杀逢印，木火通明”）并保存
- **THEN** 该名人记录被存入 Postgres 数据库的 `celebrity_records` 表，并默认处于启用池(active)中

#### Scenario: 动态停用名人
- **WHEN** 管理员点击某名人的 "下线" 按钮
- **THEN** 该条记录的 `active` 状态变为 false，大模型将在其后的请求中不再收到该名人的数据

### Requirement: AI 报告附加上下文
系统从数据库向提示词注入名人数据库快照。

#### Scenario: AI 报告请求构建
- **WHEN** 普通用户提交请求 `/api/bazi/report` 进行算盘
- **THEN** 后端服务在组织 prompt 前，检索出所有 `active=true` 的名人，将其 "姓名" 和 "特征描述" 作为数组附加到系统提示词内，强制要求大模型分析命主与这些名人的相似度并在报告尾部体现
